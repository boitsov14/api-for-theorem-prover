#!/bin/bash
ID=$1
SEQUENT=$2
cd workdir
# main.jar の実行
# 制限時間を5分に制限
# heap size を300MBに制限
# stack size を512KBに制限
timeout 300 java -Xmx300m -Xss512k -XX:CICompilerCount=2 -jar ../main.jar "$ID" "$SEQUENT" 1>"$ID"_message.txt 2>"$ID"_jar_error.txt

EXIT_STATUS=$?

# Timeoutしたとき
if [[ $EXIT_STATUS -eq 124 ]]; then
    echo -n "Proof Failed: Timeout." >>"$ID"_message.txt
    echo "Timeout"
fi

# OutOfMemoryErrorしたとき
if grep -q "OutOfMemoryError" "$ID"_jar_error.txt; then
    echo -n "Proof Failed: OutOfMemoryError." >>"$ID"_message.txt
    echo "OutOfMemoryError"
elif [[ $EXIT_STATUS -ne 0 ]]; then
    echo -n "An unexpected error has occurred: Java exec failure." >>"$ID"_message.txt
    echo "An unexpected error has occurred: Java exec failure."
fi

# ID.tex が存在しているとき
if [[ -e "$ID".tex ]]; then
    # 標準出力を ID.log に追記し，標準エラー出力を ID_error.logに追記する
    latex -halt-on-error "$ID".tex 1>>"$ID".log 2>>"$ID"_error.log
fi

# ID.dvi が存在しているとき
if [[ -e "$ID".dvi ]]; then
    # 標準出力を ID.log に追記し，標準エラー出力を ID_error.logに追記する
    dvipng "$ID".dvi 1>>"$ID".log 2>>"$ID"_error.log
fi

# ID_error.log が存在しているとき
if [[ -e "$ID"_error.log ]]; then
    # ID_error.logを標準エラー出力として表示
    cat "$ID"_error.log 1>&2

    # ID_error.logをID.logに追記
    cat "$ID"_error.log >>"$ID".log
fi
