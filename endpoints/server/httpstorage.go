package server

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/OpenNHP/opennhp/nhp/log"
)

const (
	uploadDir   = "etc/uploads"  // storage directory
	metadataDir = "etc/metadata" // metadata directory
)

type FileMetadata struct {
	UUID      string `json:"uuid"`       // file UUID
	Original  string `json:"original"`   // original file name
	MD5       string `json:"md5"`        // file MD5
	Path      string `json:"path"`       // file storage path
	Size      int64  `json:"size"`       // file size
	UploadURI string `json:"upload_uri"` // file download URI
}

// initStorageRouter mounts /storage unless http.toml disables it.
//
// Upload and metadata require a request signed by an nhp-db listed in
// db.toml (requireDBSignature), and are rate limited per source IP.
// Download stays anonymous and unlimited: agents fetch ZTDOs by the URL
// the nhp-db hands out, without credentials, and the URL carries a random
// UUID. There is no progress endpoint: its key was the UUID, which the
// client only learns once the upload has finished.
func (hs *HttpServer) initStorageRouter() {
	if !storageEnabled(hs.config) {
		log.Info("storage API disabled (EnableStorage = false)")
		return
	}
	hs.storage = newStorageState(hs.config)
	st := hs.storage

	g := hs.ginEngine.Group("/storage")
	signed := g.Group("", st.rateLimit, hs.requireDBSignature)

	signed.POST("/upload", func(c *gin.Context) {
		// Bound concurrent uploads before reading any of the body.
		select {
		case st.uploadSlots <- struct{}{}:
			defer func() { <-st.uploadSlots }()
		default:
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many uploads in progress"})
			return
		}

		// The signature covers Content-Length, so it must be known; it also
		// lets the size and free-space checks run before anything is read.
		total := c.Request.ContentLength
		if total <= 0 {
			c.JSON(http.StatusLengthRequired, gin.H{"error": "Content-Length required"})
			return
		}
		if total > st.maxUploadBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds max upload size", "max limit": st.maxUploadBytes})
			return
		}
		storeRoot := filepath.Join(ExeDirPath, uploadDir)
		if err := os.MkdirAll(storeRoot, 0o700); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create storage directory failed"})
			return
		}
		// Count the uploads already in progress against the free space too,
		// or several could each pass the check against the same space.
		reserved := st.reservedBytes.Add(total)
		defer st.reservedBytes.Add(-total)
		if free, ok := diskFreeBytes(storeRoot); ok && free-reserved < st.minFreeBytes {
			log.Warning("storage: refusing a %d-byte upload, %d bytes free on the upload volume, %d reserved by uploads in progress", total, free, reserved-total)
			c.JSON(http.StatusInsufficientStorage, gin.H{"error": "insufficient storage"})
			return
		}
		// Hash the whole body as it is read; it must match the digest the
		// nhp-db signed, or an on-path attacker could swap the content.
		bodyHash := sha256.New()
		body := http.MaxBytesReader(c.Writer, c.Request.Body, st.maxUploadBytes)
		c.Request.Body = struct {
			io.Reader
			io.Closer
		}{io.TeeReader(body, bodyHash), body}

		// Stream the file part straight to its final location. Parsing the
		// whole form first would spill large uploads to a temporary file
		// and then copy them, using twice the disk space.
		mr, err := c.Request.MultipartReader()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "multipart form expected"})
			return
		}
		var part *multipart.Part
		for {
			part, err = mr.NextPart()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "file not found"})
				return
			}
			if part.FormName() == "file" {
				break
			}
			_ = part.Close()
		}
		defer part.Close()

		filename := part.FileName()
		if !isPlainStorageName(filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
			return
		}

		// generate UUID and file path
		fileUUID := uuid.New().String()
		fileDir := filepath.Join(storeRoot, fileUUID)
		if mkdirErr := os.Mkdir(fileDir, 0o700); mkdirErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create storage directory failed"})
			return
		}
		// Remove the object directory on every failure below, so a failed
		// or rejected upload leaves nothing behind.
		stored := false
		defer func() {
			if !stored {
				_ = os.RemoveAll(fileDir)
			}
		}()
		filePath := filepath.Join(fileDir, filename)
		out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create file failed"})
			return
		}

		md5Hash := md5.New()
		written, copyErr := io.Copy(io.MultiWriter(out, md5Hash), part)
		closeErr := out.Close()
		if copyErr == nil {
			// Read the rest of the body (the closing boundary) so the
			// digest covers all of it.
			_, copyErr = io.Copy(io.Discard, c.Request.Body)
		}
		if copyErr != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(copyErr, &tooLarge) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds max upload size", "max limit": st.maxUploadBytes})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file copy failed"})
			return
		}
		if closeErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file copy failed"})
			return
		}
		if hex.EncodeToString(bodyHash.Sum(nil)) != c.GetString(storageBodyHashKey) {
			log.Warning("storage: upload from %s does not match its signed body digest", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "body does not match the signed digest"})
			return
		}

		// calculate md5
		fileMD5 := hex.EncodeToString(md5Hash.Sum(nil))

		// check if file already exists (the deferred cleanup removes this copy)
		existingMetadata, exists := checkFileExists(fileMD5)
		if exists {
			c.JSON(http.StatusOK, gin.H{
				"message":  "file already exists, skip storage",
				"file_uri": existingMetadata.UploadURI,
				"uuid":     existingMetadata.UUID,
				"md5":      existingMetadata.MD5,
			})
			return
		}

		// create metadata
		relativePath := filepath.Join(fileUUID, filename)
		fileURI := fmt.Sprintf("storage/download/%s/%s", fileUUID, filename)
		metadata := FileMetadata{
			UUID:      fileUUID,
			Original:  filename,
			MD5:       fileMD5,
			Path:      relativePath,
			Size:      written,
			UploadURI: fileURI,
		}

		// save metadata
		if err := saveMetadata(metadata); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save metadata failed"})
			return
		}
		stored = true

		log.Info("storage: %s uploaded %s (%d bytes) as %s", c.GetString(storageCallerKey), filename, written, fileUUID)
		c.JSON(http.StatusOK, gin.H{
			"message":  "file upload success",
			"file_uri": fileURI,
			"uuid":     fileUUID,
			"md5":      fileMD5,
		})
	})

	// file download
	g.GET("/download/:uuid/:filename", func(c *gin.Context) {
		uuid := c.Param("uuid")
		filename := c.Param("filename")

		// validate that uuid and filename are single path components
		if uuid == "" || strings.Contains(uuid, "/") || strings.Contains(uuid, "\\") || strings.Contains(uuid, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
			return
		}
		if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
			return
		}

		root, err := os.OpenRoot(filepath.Join(ExeDirPath, uploadDir))
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "file not exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		defer func() { _ = root.Close() }()
		// Root.OpenFile keeps resolution inside uploads even through symlinks.
		// Nonblocking open prevents a planted FIFO from holding the request.
		f, err := root.OpenFile(filepath.Join(uuid, filename), os.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "file not exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		defer func() { _ = f.Close() }()

		fileInfo, err := f.Stat()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !fileInfo.Mode().IsRegular() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
			return
		}

		// provide file download
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		c.Header("Content-Type", "application/octet-stream")
		http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), f)
	})

	// get file metadata
	signed.GET("/metadata/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		metadata, err := loadMetadata(uuid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "file metadata not exists"})
			return
		}

		c.JSON(http.StatusOK, metadata)
	})
}

// isPlainStorageName reports whether name is usable as a stored file name:
// one path element, not "." or "..", no separator of any platform and no
// control characters.
func isPlainStorageName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	for _, r := range name {
		if r == '/' || r == '\\' || r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

// saveMetadata use to save file metadata
func saveMetadata(metadata FileMetadata) error {
	if _, err := os.Stat(filepath.Join(ExeDirPath, metadataDir)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(ExeDirPath, metadataDir), os.ModePerm); err != nil {
			return err
		}
	}

	metadataPath := filepath.Join(ExeDirPath, metadataDir, metadata.UUID+".json")
	file, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

// loadMetadata use to load file metadata
func loadMetadata(uuid string) (FileMetadata, error) {
	var metadata FileMetadata
	metadataPath := filepath.Join(ExeDirPath, metadataDir, uuid+".json")

	absPath, err := filepath.Abs(metadataPath)
	if err != nil {
		return metadata, err
	}

	safeDir := filepath.Join(ExeDirPath, metadataDir)
	safeDirAbs, err := filepath.Abs(safeDir)
	if err != nil {
		return metadata, err
	}
	if !strings.HasPrefix(absPath, safeDirAbs) {
		return metadata, fmt.Errorf("invalid file name")
	}

	file, err := os.Open(absPath)
	if err != nil {
		return metadata, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&metadata)
	return metadata, err
}

// checkFileExists use to check if file exists
func checkFileExists(md5 string) (FileMetadata, bool) {
	// check all metadata files
	files, err := os.ReadDir(filepath.Join(ExeDirPath, metadataDir))
	if err != nil {
		return FileMetadata{}, false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		metadata, err := loadMetadata(file.Name()[:len(file.Name())-5]) // remove .json suffix
		if err == nil && metadata.MD5 == md5 {
			return metadata, true
		}
	}

	return FileMetadata{}, false
}
