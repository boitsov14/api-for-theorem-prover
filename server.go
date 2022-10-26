package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	app := fiber.New()

	app.Post("/twitter", func(c *fiber.Ctx) error {

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

		c.SendStatus(fiber.StatusOK)

		fmt.Println(tweet)

		processTweet(tweet)

		return nil
	})

	// portの設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
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

	msg := prove(id, sequent)
	msg += makeDVI(id, sequent)
	msg += makeImg(id, sequent)
	resizeImg(id)

	// ランダムな文字列の生成
	rand.Seed(time.Now().UnixNano())
	ran := make([]byte, 4)
	for i := range ran {
		ran[i] = byte(rand.Intn(26)%26 + 97)
	}

	text := ".@" + username + " " + msg + " [" + string(ran) + "]"

	sendTweet(id, text, username)

	// 初期ディレクトリに戻る
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}
}

func prove(id, sequent string) string {
	// proverの実行
	// 制限時間を5分に制限
	// heap sizeを300MBに制限
	// stack sizeを512KBに制限
	stdout, stderr, err := CommandExecWithTimeout(5*time.Minute, "../prover", "-Dfile.encoding=UTF-8", "-Xmx300m", "-Xss512k", id, sequent)

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

func makeDVI(id, sequent string) string {

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

func makeImg(id, sequent string) string {

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

func sendTweet(id, text, username string) {

	api := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("ACCESS_TOKEN"),
		os.Getenv("ACCESS_TOKEN_SECRET"),
		os.Getenv("API_KEY"),
		os.Getenv("API_KEY_SECRET"),
	)

	// パラメータの設定
	v := url.Values{}
	v.Add("in_reply_to_status_id", id)

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

	tweet, err := api.PostTweet(text, v)
	if err != nil {
		fmt.Println("Could not post a tweet.")
		log.Fatal(err)
	}

	fmt.Println("Tweet Succeeds:", tweet.Text)
}
