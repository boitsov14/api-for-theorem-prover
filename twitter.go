package main

import (
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

func twitterBot(c *fiber.Ctx) error {
	// 認証確認
	if c.GetReqHeaders()["Authorization"] != "Bearer "+os.Getenv("PASSWORD") {
		fmt.Println("Unauthorized request has been detected.")
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Tweetの取得
	tweet := new(TweetData)
	if err := c.BodyParser(tweet); err != nil {
		fmt.Println("Could not parse body.")
		return err
	}

	fmt.Println(tweet)

	go processTweet(tweet)

	return c.SendStatus(fiber.StatusOK)
}

type TweetData struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Username string `json:"username"`
}

func processTweet(tweet *TweetData) {

	id, sequent, username := tweet.ID, tweet.Text, tweet.Username

	// workディレクトリの作成&移動
	os.Mkdir("work", os.ModePerm)
	if err := os.Chdir("work"); err != nil {
		log.Fatal(err)
	}

	// tweetのテキストの文字列置換
	sequent = strings.ReplaceAll(sequent, "@sequent_bot", "")
	sequent = strings.ReplaceAll(sequent, "&lt;", "<")
	sequent = strings.ReplaceAll(sequent, "&gt;", ">")
	sequent = strings.ReplaceAll(sequent, "&amp;", "&")

	msg := makeProofTree(id, sequent, "2g", 1*time.Minute)

	resizeImg(id)

	sendTweet(id, msg, username)

	// 初期ディレクトリに戻る
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}
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
