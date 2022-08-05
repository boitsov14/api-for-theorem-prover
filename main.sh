#!/bin/bash
ID=$1
SEQUENT=$2
cd workdir
java -jar ../main.jar "$ID" "$SEQUENT"              # proverの実行

if [[ -e "$ID".tex ]]; then                         # ID.tex が存在していれば
    latex -halt-on-error "$ID".tex >> "$ID".log     # 標準出力を ID.log に追記する
fi

if [[ -e "$ID".dvi ]]; then                         # ファイル file が存在していれば
    dvipng "$ID".dvi >> "$ID".log                   # 標準出力を ID.log に追記する
fi
