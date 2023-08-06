# Y-III: 故障を乗り越えて動くシステムのための分散合意

## 開発環境
- 使用言語: Go 1.20
- 推奨環境: Docker + makeコマンド

## プログラムの起動方法

### Dockerを使用する場合
- `make up`: Docker Composeで開発環境を起動する
  - Server×5が同じネットワーク内に起動する
  - Client実行用コンテナが同じネットワーク内に起動する
- `make down`: `make up`で建てた開発環境を停止する
  - Server×5が停止する
  - Client実行用コンテナが停止する
- `make client`: Clientをインタラクティブモードで起動する

ネットワーク構成
| Server | Address            |
|--------|--------------------|
| node01 | 172.26.250.11:8000 |
| node02 | 172.26.250.12:8000 |
| node03 | 172.26.250.13:8000 |
| node04 | 172.26.250.14:8000 |
| node05 | 172.26.250.15:8000 |


### Dockerコマンドを使用しない場合

#### Server
コマンドオプション
```sh
> go run cmd/server/main.go --help
    -d, --delay float     Mean delay of communication channel
    -h, --host string     Host name (default "localhost")
    -l, --loss float      Loss rate of communication channel
    -n, --name string     Name of this node (default "node")
    -p, --port string     Port to listen (default "8080")
        --seed int        Random seed
    -s, --server string   Server address to join P2P network
```

起動例 (サーバ5台でシステムを構成する場合):
- Node 1: `go run cmd/server/main.go --port 8001 --name node01`
    - サーバネットワークを構成する最初の1台 (`node01`) を起動
- Node 2: `go run cmd/server/main.go --port 8002 --name node02 --server localhost:8001`
    - 新しい名前 (`node02`) でサーバを起動し，Node 1へ接続する
- Node 3: `go run cmd/server/main.go --port 8003 --name node03 --server localhost:8001`
    - 新しい名前 (`node03`) でサーバを起動し，Node 1へ接続する
- Node 4: `go run cmd/server/main.go --port 8004 --name node04 --server localhost:8001`
    - 新しい名前 (`node04`) でサーバを起動し，Node 1へ接続する
- Node 5: `go run cmd/server/main.go --port 8005 --name node05 --server localhost:8001`
    - 新しい名前 (`node05`) でサーバを起動し，Node 1へ接続する

【注意】
- 別のアドレスを持つ同名のサーバは接続できない
- Node 2~5はNode 1への接続することで自動的に相互接続されるよう設計されているが，稀に失敗する


#### Client
コマンドオプション
```sh
> go run cmd/client/main.go --help
    -e, --exec string   Execute the given command directly instead of interactive mode
```

起動例:
- インタラクティブモード: `go run cmd/client/main.go`
    - 起動すると `command ?` を表示し，入力待機状態になる．
    - コマンド例: `state localhost:8001` = `localhost:8001`に対し内部状態を問い合わせる
- コマンドの直接実行: `go run cmd/client/main.go --exec "state localhost:8001"`
    - `--exec`オプションでコマンド単発実行が可能

## 開発協力者
- [logica](https://github.com/logica0419)
