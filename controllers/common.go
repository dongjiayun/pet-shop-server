package controllers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dongjiayun/pet-shop-server/config"
	"github.com/dongjiayun/pet-shop-server/models"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

func GetObsToken(c *gin.Context) {
	type ObsTokenReq struct {
		Bucket string `json:"bucket"`
	}
	var req ObsTokenReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: err.Error()})
		return
	}
	ch := make(chan string)
	go models.GetObsToken(req.Bucket, ch)
	obsToken := <-ch
	c.JSON(200, models.Result{0, "success", obsToken})
}

func GetAllArea(c *gin.Context) {
	type Area struct {
		Id       int    `json:"id"`
		Code     string `json:"code"`
		Name     string `json:"name"`
		Children []Area `json:"children,omitempty"`
	}
	var area []Area
	json.Unmarshal([]byte(config.AreaDict), &area)
	c.JSON(200, models.Result{0, "success", area})
}

const maxUploadSize = 2 << 20 // 2MB in bytes

func CommonUploadPic(c *gin.Context) {
	// 1. Get the file from the form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(200, models.Result{Code: 10001, Message: "获取文件失败: " + err.Error()})
		return
	}

	// 2. Check file size (max 2MB)
	if file.Size > maxUploadSize {
		c.JSON(200, models.Result{Code: 10011, Message: "文件大小不能超过2MB"})
		return
	}

	// 3. Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	if !allowedExtensions[ext] {
		c.JSON(200, models.Result{Code: 10004, Message: "不支持的文件类型，仅支持: jpg, jpeg, png, gif, bmp, webp"})
		return
	}

	// 4. Open and validate file
	fileHeader, err := file.Open()
	if err != nil {
		c.JSON(200, models.Result{Code: 10005, Message: "文件打开失败: " + err.Error()})
		return
	}
	defer fileHeader.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := fileHeader.Read(buffer)
	if err != nil && err != io.EOF {
		c.JSON(200, models.Result{Code: 10006, Message: "文件读取失败: " + err.Error()})
		return
	}

	// Check MIME type
	contentType := http.DetectContentType(buffer[:n])
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(200, models.Result{Code: 10007, Message: "无效的图片文件"})
		return
	}

	// 5. Create uploads directory if not exists
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(uploadDir, 0755)
		if err != nil {
			c.JSON(200, models.Result{Code: 10002, Message: "创建上传目录失败: " + err.Error()})
			return
		}
	}

	// 6. Generate UUID filename and save file
	newFilename := uuid.New().String() + ext
	dst := filepath.Join(uploadDir, newFilename)

	// Reset file reader to the beginning
	if _, err = fileHeader.Seek(0, 0); err != nil {
		c.JSON(200, models.Result{Code: 10008, Message: "文件处理失败: " + err.Error()})
		return
	}

	// Create the file
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(200, models.Result{Code: 10009, Message: "创建文件失败: " + err.Error()})
		return
	}
	defer out.Close()

	// Copy the file content
	_, err = io.Copy(out, fileHeader)
	if err != nil {
		// Clean up the file if copy fails
		os.Remove(dst)
		c.JSON(200, models.Result{Code: 10010, Message: "保存文件失败: " + err.Error()})
		return
	}

	// 7. Return the file URL/path
	fileURL := "/uploads/" + newFilename
	c.JSON(200, models.Result{
		Code:    0,
		Message: "文件上传成功",
		Data:    fileURL,
	})
}
