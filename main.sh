#!/bin/bash
ID=$1
SEQUENT=$2
cd workdir
# main.jar の実行
# heap size を300MBに制限
# stack size を512KBに制限
java -Xmx300m -Xss512k -XX:CICompilerCount=2 -jar ../main.jar "$ID" "$SEQUENT" > "$ID"_message.txt

# ID.tex が存在しているとき
if [[ -e "$ID".tex ]]; then
    # 標準出力を ID.log に追記し，標準エラー出力を ID_error.logに追記する
    latex -halt-on-error "$ID".tex 1>> "$ID".log 2>> "$ID"_error.log
fi

# ID.dvi が存在しているとき
if [[ -e "$ID".dvi ]]; then
    # 標準出力を ID.log に追記し，標準エラー出力を ID_error.logに追記する
    dvipng "$ID".dvi 1>> "$ID".log 2>> "$ID"_error.log
fi

# ID_error.logを標準エラー出力として表示
cat "$ID"_error.log 1>&2

# ID_error.logをID.logに追記
cat "$ID"_error.log >> "$ID".log
