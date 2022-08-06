'use strict'

require('dotenv').config()
const express = require('express')
const fs = require('fs')
const { exec } = require('child_process')
const sizeOf = require('image-size')
const sharp = require('sharp')
const { TwitterApi } = require('twitter-api-v2')
const client = new TwitterApi({
    appKey: process.env.API_KEY,
    appSecret: process.env.API_KEY_SECRET,
    accessToken: process.env.ACCESS_TOKEN,
    accessSecret: process.env.ACCESS_TOKEN_SECRET,
})

const app = express()
const port = process.env.PORT || 5000

app.use(express.json())

app.listen(port, () => {
    console.log(`Listening on port: ${port}`)
})

app.post('/twitter_bot', (req, res) => {

    //パスワードのチェック
    if (req.body.password !== process.env.PASSWORD) {
        return res.sendStatus(400)
    }

    //GASに返信
    res.sendStatus(200)

    // tweetsの取得
    const { tweets } = req.body

    //各tweetの処理
    for (const tweet of tweets) {
        try {
            process_tweet(tweet)
        } catch (err) {
            console.log(`ERROR: ${err.message}`)
        }
    }
})

const process_tweet = async (tweet) => {

    const { id, text, username } = tweet

    //tweetのテキストの文字列置換
    const sequent = text
        .replaceAll('@sequent_bot', '')
        .replace(/\s+/g, ' ')
        .replaceAll('&lt;', '<')
        .replaceAll('&gt;', '>')
        .replaceAll('&amp;', '&')

    console.log({
        'id': id,
        'sequent': sequent
    })

    //main.sh のコマンド実行
    exec(`bash main.sh "${id}" "${sequent}"`, { timeout: 10 * 60 * 1000 }, (error, stdout, stderr) => {
        if (error) {
            console.log(`BASH ERROR: ${error}`)
        }
        if (stderr) {
            console.log(`BASH STDERR: ${stderr}`)
        }
        if (stdout) {
            console.log(`BASH STDOUT: ${stdout}`)
        }
        process_tweet_core(id, username)
    })
}

const process_tweet_core = async (id, username) => {
    //tweet文章と画像
    let message = ''
    let image = ''

    //message.txt の読み込み
    if (!(fs.existsSync(`./workdir/${id}_message.txt`))) {
        //main.jar がmessageファイルに書き込む前に異常終了したとき
        message = 'An unexpected error has occurred: No message file.'
    } else {
        message = fs.readFileSync(`./workdir/${id}_message.txt`, 'utf-8')
    }

    //Texファイルが存在するとき
    if (fs.existsSync(`./workdir/${id}.tex`)) {
        if (!(fs.existsSync(`./workdir/${id}.log`))) {
            //logファイルが出力される前に異常終了が起きたとき
            message += 'An unexpected error has occurred: No tex log file.'
        } else {
            //logファイルの読み込み
            const log = fs.readFileSync(`./workdir/${id}.log`, 'utf-8')

            if (log.includes('Dimension too large')) {
                //Dimension too largeのとき
                message += 'The proof tree is too large to output: Dimension too large.'
            } else if (log.includes('DVI stack overflow')) {
                //Fatal error, DVI stack overflowのとき
                message += 'The proof tree is too large to output: DVI stack overflow.'
            } else if (!(fs.existsSync(`./workdir/${id}1.png`))) {
                //その他の予期せぬ理由によりPNGが生成されないとき
                message += 'An unexpected error has occurred: No png file.'
            }
        }
    }

    //PNGが存在するとき
    if (fs.existsSync(`./workdir/${id}1.png`)) {
        const size = sizeOf(`./workdir/${id}1.png`)

        //画像サイズが大きすぎるときは縮小する
        if (size.width > 8192 || size.height > 8192) {
            console.log(`The image size is too large: ${size.width} * ${size.height}`)
            await sharp(`./workdir/${id}1.png`)
                .resize(8192, 8192, { fit: 'inside' })
                .toFile(`./workdir/${id}1.png`)
        }
        image = `./workdir/${id}1.png`
    }

    //.@ユーザ名 + message + 長さが4~5のランダムな文字列
    const tweet_text = `.@${username} ${message} (${Math.random().toString(32).substring(8)})`

    //ツイートする
    //await make_reply(id, tweet_text, image)

    console.log(`Tweet Success: ${tweet_text}`)
}

const make_reply = async (id, message, image) => {
    if (image) {
        //image付きツイート
        const media_id = await client.v1.uploadMedia(image)
        await client.v1.tweet(
            message, {
            in_reply_to_status_id: id,
            media_ids: media_id
        })
    } else {
        //image無しツイート
        await client.v1.tweet(
            message, {
            in_reply_to_status_id: id
        })
    }
}
