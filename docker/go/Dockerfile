# Goの公式イメージを使用
FROM golang:1.24.3-alpine

# 作業ディレクトリを作成
WORKDIR /go/src/app/

# パッケージをインストール
RUN apk update && apk add git

# アプリケーションコードをコンテナ内にコピー
COPY ./src/ ./

# モジュールをインストール
RUN go mod download

RUN go build -o /go/bin/app .

# デフォルトでポート 8080 を使用
ENV PORT=8080

# コンテナがリッスンするポートを指定
EXPOSE 8080

# 実行コマンドを指定
CMD ["/go/bin/app"]