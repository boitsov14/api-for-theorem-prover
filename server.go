package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	app := fiber.New()

	app.Post("/twitter", postTwitter)
	app.Post("/web", postWeb)

	// portの設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
}

func postTwitter(c *fiber.Ctx) error {
	// 認証確認
	if c.GetReqHeaders()["Authorization"] != "Bearer "+os.Getenv("PASSWORD") {
		fmt.Println("Unauthorized request has been detected.")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Tweetの取得
	tweet := new(Tweet)
	if err := c.BodyParser(tweet); err != nil {
		fmt.Println("Could not parse body.")
		return err
	}

	fmt.Println(tweet)

	go processTweet(tweet)

	return c.SendStatus(fiber.StatusOK)
}

type Tweet struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Username string `json:"username"`
}

func processTweet(tweet *Tweet) {

	id, sequent, username := tweet.ID, tweet.Text, tweet.Username

	// workディレクトリの作成&移動
	os.Mkdir("work", os.ModePerm)
	if err := os.Chdir("work"); err != nil {
		log.Fatal(err)
	}

	// tweetのテキストの文字列置換
	sequent = strings.ReplaceAll(sequent, "@sequent_bot", "")
	sequent = strings.Join(strings.Fields(sequent), " ")
	sequent = strings.ReplaceAll(sequent, "&lt;", "<")
	sequent = strings.ReplaceAll(sequent, "&gt;", ">")
	sequent = strings.ReplaceAll(sequent, "&amp;", "&")

	msg := prove(id, sequent, "2g", 1*time.Minute)
	msg += makeDVI(id)
	msg += makeImg(id)
	resizeImg(id)

	sendTweet(id, msg, username)

	// 初期ディレクトリに戻る
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}
}

func postWeb(c *fiber.Ctx) error {

	// Bodyの取得
	request := new(Request)
	if err := c.BodyParser(request); err != nil {
		fmt.Println("Could not parse body.")
		return err
	}

	fmt.Println(request)
	//TODO: Line notify

	return c.JSON(processRequest(request.Text))
}

type Request struct {
	Text string `json:"text"`
}

type Response struct {
	Message string `json:"message"`
	Image   string `json:"image"`
	Tex     string `json:"tex"`
}

func processRequest(sequent string) *Response {

	response := new(Response)

	id := uuid.NewString()

	// workディレクトリの作成&移動
	os.Mkdir("work", os.ModePerm)
	if err := os.Chdir("work"); err != nil {
		log.Fatal(err)
	}

	// tweetのテキストの文字列置換
	// TODO これは必要なのか
	sequent = strings.Join(strings.Fields(sequent), " ")

	msg := prove(id, sequent, "300m", 10*time.Second)
	msg += makeDVI(id)
	msg += makeImg(id)

	response.Message = msg
	//TODO: Line notify

	if exists(id + ".png") {
		// Image
		imgBytes, err := os.ReadFile(id + ".png")
		if err != nil {
			log.Fatal(err)
		}
		response.Image = base64.StdEncoding.EncodeToString(imgBytes)
		//TODO: Line notify
		// Tex
		texBytes, err := os.ReadFile(id + ".tex")
		if err != nil {
			log.Fatal(err)
		}
		response.Tex = string(texBytes)
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

func prove(id, sequent, size string, timeout time.Duration) string {

	stdout, stderr, err := CommandExecWithTimeout(timeout, "../prover", "-Xmx"+size, id, sequent)

	// Timeoutしたとき
	if err == context.DeadlineExceeded {
		if stdout == "" {
			return "Proof Failed: Timeout."
		} else {
			return stdout + " The proof tree is too large to output: Timeout."
		}
	}

	// OutOfMemoryErrorしたとき
	if strings.Contains(stderr, "OutOfMemoryError") {
		if stdout == "" {
			return "Proof Failed: OutOfMemoryError."
		} else {
			return stdout + " The proof tree is too large to output: OutOfMemoryError."
		}
	}

	// その他のエラーが発生したとき
	if stdout == "" || stderr != "" || err != nil {
		fmt.Println(stdout, stderr, err)
		return stdout + "An unexpected error has occurred: Java exec failure."
	}

	return stdout
}

func makeDVI(id string) string {

	if !exists(id + ".tex") {
		return ""
	}

	stdout, stderr, err := CommandExec("latex", "-halt-on-error", id+".tex")

	// Dimension too largeのとき
	if strings.Contains(stdout, "Dimension too large") {
		return " The proof tree is too large to output: Dimension too large."
	}

	// その他の理由によりDVIが生成されないとき
	if !exists(id + ".dvi") {
		fmt.Println(stdout, stderr, err)
		return " An unexpected error has occurred: Could not compile tex file."
	}

	return ""
}

func makeImg(id string) string {

	if !exists(id + ".dvi") {
		return ""
	}

	stdout, stderr, err := CommandExec("dvipng", id+".dvi", "-o", id+".png")

	// DVI stack overflowのとき
	if strings.Contains(stderr, "DVI stack overflow") {
		return " The proof tree is too large to output: DVI stack overflow."
	}

	// その他の理由によりPNGが生成されないとき
	if !exists(id + ".png") {
		fmt.Println(stdout, stderr, err)
		return " An unexpected error has occurred: Could not compile dvi file."
	}

	return ""
}

func sendTweet(id, msg, username string) {

	api := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("ACCESS_TOKEN"),
		os.Getenv("ACCESS_TOKEN_SECRET"),
		os.Getenv("API_KEY"),
		os.Getenv("API_KEY_SECRET"),
	)

	// Tweetするテキスト
	text := ".@" + username + " " + msg

	// パラメータの設定
	v := url.Values{}
	v.Add("in_reply_to_status_id", id)

	// 画像のアップロード処理
	if exists(id + ".png") {
		// PNGをBASE64に変換
		bytes, err := os.ReadFile(id + ".png")
		if err != nil {
			log.Fatal(err)
		}
		base64img := base64.StdEncoding.EncodeToString(bytes)
		// Twitterに画像をアップロード
		media, err := api.UploadMedia(base64img)
		if err != nil {
			fmt.Println("Could not upload media.")
			log.Fatal(err)
		}
		v.Add("media_ids", media.MediaIDString)
	}

	// ランダム文字列の設定
	if !exists(id+".png") && !strings.Contains(msg, "seconds") {
		// 長さ3のランダムな文字列の生成
		rand.Seed(time.Now().UnixNano())
		ran := make([]byte, 3)
		for i := range ran {
			ran[i] = byte(rand.Intn(26)%26 + 97)
		}
		text += " [" + string(ran) + "]"
	}

	tweet, err := api.PostTweet(text, v)
	if err != nil {
		fmt.Println("Could not post a tweet.")
		log.Fatal(err)
	}

	fmt.Println("Tweet Succeeds:", tweet.Text)
}
