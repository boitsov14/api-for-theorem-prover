package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func webApp(c *fiber.Ctx) error {

	// Bodyの取得
	request := new(WebRequest)
	if err := c.BodyParser(request); err != nil {
		fmt.Println("Could not parse body.")
		return err
	}

	fmt.Println(request.Text)

	// Line通知
	notifyLine(request.Text)

	return c.JSON(processWebRequest(request.Text))
}

type WebRequest struct {
	Text string `json:"text"`
}

type WebResponse struct {
	Message string `json:"message"`
	Image   string `json:"image"`
	Tex     string `json:"tex"`
}

func processWebRequest(sequent string) *WebResponse {

	response := new(WebResponse)

	id := uuid.NewString()

	// workディレクトリの作成&移動
	os.Mkdir("work", os.ModePerm)
	if err := os.Chdir("work"); err != nil {
		log.Fatal(err)
	}

	msg := makeProofTree(id, sequent, "200m", 10*time.Second)

	fmt.Println(msg)
	response.Message = msg

	if exists(id + ".png") {
		// Line通知
		notifyLineWithProofTree(msg, id)

		// Image
		imgBytes, err := os.ReadFile(id + ".png")
		if err != nil {
			log.Fatal(err)
		}
		response.Image = base64.StdEncoding.EncodeToString(imgBytes)

		// Tex
		texBytes, err := os.ReadFile(id + ".tex")
		if err != nil {
			log.Fatal(err)
		}
		response.Tex = string(texBytes)
	} else {
		// Line通知
		notifyLine(msg)
	}

	// 生成したファイルの削除
	files, err := filepath.Glob(id + ".*")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Fatal(err)
		}
	}

	// 初期ディレクトリに戻る
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}

	return response
}
