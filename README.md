# prover-twitter-bot-docker

論理式をリプライで送ると，証明図を返してくれる Twitter bot([@sequent_bot](https://twitter.com/sequent_bot))

## 使い方

- `.env`ファイルを作成し，以下の内容を書く(事前に Twitter の API KEY 等を準備しておく)．

```
API_KEY=<TwitterのAPI KEY>
API_KEY_SECRET=<TwitterのAPI KEY SECRET>
ACCESS_TOKEN=<TwitterのACCESS TOKEN>
ACCESS_TOKEN_SECRET=<TwitterのACCESS TOKEN SECRET>
PASSWORD=<なんか適当な文字列>
```

- [theorem-prover-kt](https://github.com/boitsov14/theorem-prover-kt/releases/tag/v1.0.0)の jar ファイルを`prover.jar`という名前で配置
- 次を実行し，Docker image を作成

```
docker build -t <付けたいimageの名前> .
```

- 次を実行し，Docker image を実行する

```
docker run -dp 3000:3000 <上で付けたimageの名前>
```

- 以下のような形の JSON を`localhost:3000/twitter`に POST リクエストする．ただし，上で設定したパスワードを Bearer トークンとしてヘッダー内に記述すること．

```
{
    "username": <返信したい相手のユーザー名>,
    "id": <返信したいツイートのid>,
    "text": <ツイートの内容>
}

```

## Heroku へのデプロイ

- Heroku のアカウントを作成し，新しいアプリを作成する

- コマンドで以下を実行

```
# Herokuへのログイン
heroku login

# Herokuコンテナへのログイン
heroku container:login

# 既存のimageをもとにHeroku用のimageを作る
docker tag <image名> registry.heroku.com/<image名>/web

# imageをHerokuにpushする
docker push registry.heroku.com/<image名>/web

# pushしたimageをリリース
heroku container:release web -a <アプリ名>
```

- GAS 等を使ってリプライが来ていないか毎分 Twitter に確認する．
- リプライが来ていた場合，上記のような形の JSON を`https://<アプリ名>.herokuapp.com/twitter`に POST リクエストする．
