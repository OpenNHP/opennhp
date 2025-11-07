//
//  FileCopyManager.swift
//  TestXCFrameworkSwift
//
//  Created by haochangjiu on 2025/10/30.
//

import UIKit
import Foundation

class FileCopyManager {

    /// Copy specified files to the etc and certs directories in the sandbox
    static func copyFilesToSandboxEtc() {
        // 1. Get the Documents directory in the sandbox
        guard let documentsURL = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first else {
            print("Failed to get Documents directory")
            return
        }

        // 2. Define paths for etc and certs directories in the sandbox
        let etcURL = documentsURL.appendingPathComponent("etc")
        let certsURL = etcURL.appendingPathComponent("certs")

        // 3. Create etc and certs directories (if they don't exist)
        createDirectoryIfNotExists(at: etcURL)
        createDirectoryIfNotExists(at: certsURL)

        // 4. Copy toml files to the etc directory
        let tomlFiles = ["server.toml", "config.toml", "dhp.toml", "resource.toml"]
        tomlFiles.forEach { fileName in
            copyFileFromBundle(fileName: fileName, to: etcURL)
        }

        // 5. Copy certificate files to the etc/certs directory
        let certFiles = ["server.crt", "server.key"]
        certFiles.forEach { fileName in
            copyFileFromBundle(fileName: fileName, to: certsURL)
        }
    }

    /// Create directory if it doesn't exist
    private static func createDirectoryIfNotExists(at url: URL) {
        let fileManager = FileManager.default
        guard !fileManager.fileExists(atPath: url.path) else {
            print("Directory already exists: \(url.path)")
            return
        }

        do {
            try fileManager.createDirectory(at: url, withIntermediateDirectories: true, attributes: nil)
            print("Directory created successfully: \(url.path)")
        } catch {
            print("Failed to create directory: \(url.path), error: \(error.localizedDescription)")
        }
    }

    /// Copy file from Bundle to destination path
    private static func copyFileFromBundle(fileName: String, to destinationURL: URL) {
        // Split filename and extension (handling files with extensions)
        let fileNameWithoutExt = (fileName as NSString).deletingPathExtension
        let fileExt = (fileName as NSString).pathExtension

        // Get the file path in the Bundle
        guard let sourceURL = Bundle.main.url(forResource: fileNameWithoutExt, withExtension: fileExt) else {
            print("File not found in Bundle: \(fileName)")
            return
        }

        // Destination file path (destination directory + filename)
        let destFileURL = destinationURL.appendingPathComponent(fileName)
        let fileManager = FileManager.default

        // Copy file (if it doesn't exist)
        guard !fileManager.fileExists(atPath: destFileURL.path) else {
            print("File already exists: \(destFileURL.path)")
            return
        }

        do {
            try fileManager.copyItem(at: sourceURL, to: destFileURL)
            print("File copied successfully: \(fileName) -> \(destFileURL.path)")
        } catch {
            print("File copy failed: \(fileName), error: \(error.localizedDescription)")
        }
    }
}
