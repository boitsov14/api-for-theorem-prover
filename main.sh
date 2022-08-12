#!/bin/bash
ID=$1
SEQUENT=$2
cd workdir

# main.jar の実行
# 制限時間を5分に制限
# heap size を300MBに制限
# stack size を512KBに制限
timeout 300 java -Xmx300m -Xss512k -XX:CICompilerCount=2 -jar ../main.jar "$ID" "$SEQUENT" 1>"$ID"_msg.txt 2>"$ID"_java_err.txt

EXIT_STATUS=$?

if [[ "$EXIT_STATUS" -eq 124 ]]; then
    # Timeoutしたとき
    if [[ -s "$ID"_msg.txt ]]; then
        # ID_msg.txtが空でないとき
        echo -n " The proof tree is too large to output: Timeout." >>"$ID"_msg.txt
    else
        # ID_msg.txtが空のとき
        echo -n "Proof Failed: Timeout." >"$ID"_msg.txt
    fi
elif grep -q "OutOfMemoryError" "$ID"_java_err.txt; then
    # OutOfMemoryErrorしたとき
    if [[ -s "$ID"_msg.txt ]]; then
        # ID_msg.txtが空でないとき
        echo -n " The proof tree is too large to output: OutOfMemoryError." >>"$ID"_msg.txt
    else
        # ID_msg.txtが空のとき
        echo -n "Proof Failed: OutOfMemoryError." >"$ID"_msg.txt
    fi
elif [[ "$EXIT_STATUS" -ne 0 ]]; then
    # 上記以外の予期せぬエラーが発生したとき
    echo -n "An unexpected error has occurred: Java exec failure." >>"$ID"_msg.txt
fi

# ID.tex が存在しているとき
if [[ -e "$ID".tex ]]; then
    # 標準出力を ID.log に追記し，標準エラー出力を ID_err.logに追記する
    latex -halt-on-error "$ID".tex 1>>"$ID".log 2>>"$ID"_err.log
    if grep -q "Dimension too large" "$ID".log; then
        # Dimension too largeのとき
        echo -n " The proof tree is too large to output: Dimension too large." >>"$ID"_msg.txt
    elif [[ ! -e "$ID".dvi ]]; then
        # その他の予期せぬ理由によりdviファイルが生成されないとき
        echo -n " An unexpected error has occurred: Could not compile tex file." >>"$ID"_msg.txt
    fi
fi

# ID.dvi が存在しているとき
if [[ -e "$ID".dvi ]]; then
    # 標準出力を ID.log に追記し，標準エラー出力を ID_err.logに追記する
    dvipng "$ID".dvi 1>>"$ID".log 2>>"$ID"_err.log
    if grep -q "DVI stack overflow" "$ID"_err.log; then
        # DVI stack overflowのとき
        echo -n " The proof tree is too large to output: DVI stack overflow." >>"$ID"_msg.txt
    elif [[ ! -e "$ID"1.png ]]; then
        # その他の予期せぬ理由によりpngファイルが生成されないとき
        echo -n " An unexpected error has occurred: Could not compile dvi file." >>"$ID"_msg.txt
    fi
fi

# ID_err.log が存在しているとき
if [[ -e "$ID"_java_err.txt ]]; then
    # ID_java_err.txtを標準エラー出力として表示
    cat "$ID"_java_err.txt 1>&2
fi

# ID_err.log が存在しているとき
if [[ -e "$ID"_err.log ]]; then
    # ID_err.logを標準エラー出力として表示
    cat "$ID"_err.log 1>&2
fi
