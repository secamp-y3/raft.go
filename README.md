# Y-III: 故障を乗り越えて動くシステムのための分散合意

## 必要環境

- Docker (Docker Compose)
- make

## 開発の進め方

- `make up`
  - Docker Composeで開発環境を建てる
  - Dispatcher×1、Console×1、Peer×5が同じネットワーク内に建つ
- `make down`
  - `make up`で建てた開発環境を削除する
- `make console`
  - 分散システムを操作するためのコンソールを起動する
