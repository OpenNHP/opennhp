package test

import (
	"net/url"
	"testing"
)

/*
func TestZLBTokenAPIOld(t *testing.T) {
	accessKey := "wxtyjrqsmzqglbzpt"
	secretKey := "wxtyjrqsmzqglbzptpwd"
	token := "8a1189bf8caba36b018cae241546098c-commonToken"
	u := "https://appapi.zjzwfw.gov.cn/sso/servlet/simpleauth?method=getUserInfo&"

	timestamp := time.Now().Format("20060102150405")
	sign := fmt.Sprintf("%s%s%s", accessKey, secretKey, timestamp)
	fmt.Printf("sign string: %s\n", sign)
	signHash := utils.MD5(sign)

	params := url.Values{}
	params.Add("servicecode", accessKey)
	params.Add("time", timestamp)
	params.Add("sign", signHash)
	params.Add("token", token)
	params.Add("datatype", "json")
	rawQuery := params.Encode()

	var headers = make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Content-Length"] = strconv.Itoa(len(rawQuery))

	resp, err := utils.Request(u, "POST", rawQuery, headers)
	fmt.Printf("Post to %s, headers: %s\ncontent:\n%s\n", u, headers, rawQuery)
	fmt.Printf("get zlb user info response: %s, err %s\n", resp, err)
}

func TestZLBTokenAPI(t *testing.T) {
	accessKey := "BCDSGA_9754b3458c581ef0efb7b66946c9c324"
	secretKey := "BCDSGS_d8d2e2f21c2a02c09a641a11cf45d91e"
	appId := "2001832595"
	var accessToken string
	zlbServerBaseUrl := "https://ibcdsg.zj.gov.cn:8443"
	getTokenUrl := "/restapi/prod/IC33000020220329000007/uc/sso/access_token"

	content := fmt.Sprintf("{\"ticketId\":\"%s\",\"appId\":\"%s\"}", "a1d67776ce344be1a240bbacd8c3daf3", appId)

	timestamp := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 MST")
	//timestamp = strings.Replace(timestamp, "UTC", "GMT", 1)
	fmt.Println("timestamp: ", timestamp)

	signStr := fmt.Sprintf("POST\n%s\n%s\n%s\n%s\n", getTokenUrl, "", accessKey, timestamp)
	//signHex := hex.EncodeToString(utils.HMACSha256(secretKey, signStr))
	signHex := utils.HMACSha256(secretKey, signStr)
	fmt.Println("sign hex: ", signHex)
	signBase64 := base64.StdEncoding.EncodeToString([]byte(signHex))
	fmt.Println("sign base64: ", signBase64)

	var headers = make(map[string]string)
	headers["Content-Type"] = "application/json"
	//headers["Content-Length"] = strconv.Itoa(len(content))
	headers["X-BG-DATE-TIME"] = timestamp
	headers["X-BG-HMAC-ACCESS-KEY"] = accessKey
	headers["X-BG-HMAC-ALGORITHM"] = "hmac-sha256"
	headers["X-BG-HMAC-SIGNATURE"] = signBase64

	resp, err := utils.Request(zlbServerBaseUrl+getTokenUrl, "POST", content, headers)
	if err != nil {
		return
	}
	fmt.Println("get access token response: ", resp)

	d := &zlb.ZLBApiResp{}
	if err = json.Unmarshal([]byte(resp), d); err != nil {
		return
	}

	if !d.Success {
		return
	}

	accessToken, ok := d.Data["accessToken"].(string)
	if !ok {
		fmt.Println("access token is not a string ")
	}
	fmt.Println("access token: ", accessToken)
}
*/

func TestUrlEncoding(t *testing.T) {
	str := "中文"
	encoded := url.QueryEscape(str)
	println("encoded: ", encoded)
	decoded, err := url.QueryUnescape(str)
	println("decoded: ", decoded, " error: ", err)
}
