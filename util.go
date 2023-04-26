package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/juunini/simple-go-line-notify/notify"
	"golang.org/x/image/draw"
)

func CommandExecWithTimeout(timeout time.Duration, command string, args ...string) (string, string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if ctx.Err() != nil {
		return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), ctx.Err()
	} else {
		return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
	}
}

func CommandExec(command string, args ...string) (string, string, error) {

	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func exists(path string) bool {

	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	return err == nil
}

func resizeImg(id string) {

	path := id + ".png"

	if !exists(path) {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	rct := img.Bounds()
	width := rct.Dx()
	height := rct.Dy()
	limit := 8192

	if width <= limit {
		return
	}

	height = height * limit / width

	dst := image.NewRGBA(image.Rect(0, 0, limit, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, rct, draw.Over, nil)

	newFile, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	png.Encode(newFile, dst)
}

func notifyLine(msg string) {
	if err := notify.SendText(os.Getenv("LINE_ACCESS_TOKEN"), msg); err != nil {
		fmt.Println("LINE Notification error: ", err)
	}
}

func notifyLineWithProofTree(msg, id string) {
	if err := notify.SendLocalImage(os.Getenv("LINE_ACCESS_TOKEN"), msg, id+".png"); err != nil {
		fmt.Println("LINE Notification error: ", err)
	}
}
