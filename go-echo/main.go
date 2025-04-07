package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yitsushi/go-misskey"
	"github.com/yitsushi/go-misskey/core"
	"github.com/yitsushi/go-misskey/models"
	"github.com/yitsushi/go-misskey/services/drive/files"
	"github.com/yitsushi/go-misskey/services/notes"
)

var (
	mutex sync.Mutex // Mutex for exclusive control of execution of prover
)

func main() {
	// create echo instance
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	e.POST("/web", func(c echo.Context) error {
		// get request
		req := new(WebReq)
		if err := c.Bind(req); err != nil {
			notify("/web: Invalid request.")
			return err
		}
		notify("/web: " + req.Txt)

		// prove
		result, err := prove(req.Txt, "500m", 10, true)
		if err != nil {
			notify("Unexpected error has occurred.\n" + err.Error())
			return c.String(http.StatusInternalServerError, "")
		}

		// create response
		res := &WebRes{
			Msg: result.Msg,
			Img: base64.StdEncoding.EncodeToString(result.Img),
			Tex: result.Tex,
		}

		// return response
		return c.JSON(http.StatusOK, res)
	})

	e.POST("/misskey", func(c echo.Context) error {
		// check password
		if c.Request().Header.Get("Authorization") != "Bearer "+os.Getenv("PASSWORD") {
			notify("/misskey: Invalid password.")
			return c.String(http.StatusUnauthorized, "")
		}

		// get request
		req := new(MisskeyReq)
		if err := c.Bind(req); err != nil {
			notify("/misskey: Invalid request.")
			return err
		}
		notify("/misskey: " + req.Txt)

		// prove
		result, err := prove(req.Txt, "2g", 30, true)
		if err != nil {
			notify("Unexpected error has occurred.\n" + err.Error())
			return c.String(http.StatusInternalServerError, "")
		}

		// add username
		result.Msg = req.Username + " " + result.Msg

		// add random string if not contains "seconds"
		if !strings.Contains(result.Msg, "seconds") {
			alphabet := "abcdefghijklmnopqrstuvwxyz"
			ran := make([]byte, 3)
			for i := range ran {
				ran[i] = alphabet[rand.Intn(len(alphabet))]
			}
			result.Msg += " [" + string(ran) + "]"
		}

		// create note
		if err := createNote(result, req.Id); err != nil {
			notify("Could not create post.\n" + err.Error())
			return c.String(http.StatusInternalServerError, "")
		}

		return c.String(http.StatusOK, "")
	})

	// get port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// start server
	e.Logger.Fatal(e.Start(":" + port))
}

type WebReq struct {
	Txt string `json:"txt"`
}

type WebRes struct {
	Msg string `json:"msg"`
	Img string `json:"img"`
	Tex string `json:"tex"`
}

type MisskeyReq struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Txt      string `json:"txt"`
}

func createNote(result *Result, renoteID string) error {
	// create misskey client
	client, err := misskey.NewClientWithOptions(misskey.WithSimpleConfig("https://misskey.io", os.Getenv("MISSKEY_ACCESS_TOKEN")))
	if err != nil {
		return err
	}

	// create request
	req := notes.CreateRequest{
		Text:       core.NewString(result.Msg),
		RenoteID:   core.NewString(renoteID),
		Visibility: models.VisibilityHome,
	}

	if result.Img != nil {
		// upload image
		file, err := client.Drive().File().Create(files.CreateRequest{
			Content: result.Img,
		})
		if err != nil {
			return err
		}

		// set file id
		req.FileIDs = []string{file.ID}

		// create note
		res, err := client.Notes().Create(req)
		if err != nil {
			return err
		}
		notify("https://misskey.io/notes/" + res.CreatedNote.ID)
	} else {
		// create note
		res, err := client.Notes().Create(req)
		if err != nil {
			return err
		}
		notify("https://misskey.io/notes/" + res.CreatedNote.ID)
	}
	return nil
}

func notify(msg string) {
	fmt.Println(msg)
	_, err := resty.
		New().
		SetRetryCount(3).
		R().
		SetBody(msg).
		SetHeader("Content-Type", "text/plain").
		Post(os.Getenv("NOTIFICATION_URL") + "/text")
	if err != nil {
		fmt.Println("Notification Error: ", err)
	}
}

func notifyWithImage(msg string) {
	fmt.Println(msg)
	_, err := resty.
		New().
		SetRetryCount(3).
		R().
		SetFormData(map[string]string{
			"content": msg,
		}).
		SetFile("file", "out.png").
		Post(os.Getenv("NOTIFICATION_URL") + "/png")
	if err != nil {
		fmt.Println("Notification Error: ", err)
	}
}
