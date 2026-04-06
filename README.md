# **TDUEx**

東京電機大学の LMS から講義一覧やイベント情報を取得して、  
複数形式で export する CLI ツールです。

## **Overview**

- LMS の時間割や講義ページを手作業で確認する手間を減らすために作成しました。
- 講義一覧だけを取るモードと、event まで辿って取得するモードを切り替えられます。
- export 形式は `JSON` / `CSV` / `XLSX` / `ICS` に対応しています。
- macOS では Finder 系、Windows では Explorer 系の保存ダイアログから保存先を選べます。

## **Architecture**

### **Tech Stack**

- Backend: Go
- Scraping: Colly / goquery
- Export: JSON / CSV / XLSX / ICS
- Config: `.env` / `.setting`
- Install: Makefile / shell script

## **Project Structure**

```text
.
├── cmd
│   └── tduex
├── pkg
│   ├── appconfig
│   ├── logger
│   ├── parser
│   ├── scraping
│   └── service
├── scripts
│   └── install.sh
├── Makefile
├── go.mod
└── README.md
```

## **Main Flow**

`tduex` が行う基本処理は以下です。

1. 設定ファイルを読み込む
2. 必要なら `USER_ID` と `PASSWORD` を入力して `.setting` に保存する
3. LMS から講義一覧または event 情報を取得する
4. 指定形式で export する

## **Commands**

### **コマンドの役割**

- `tduex classes`
  講義一覧だけを取得して export
- `tduex full`
  講義ごとの event まで辿って export

### **対話実行**

```bash
tduex
```

またはビルド前なら:

```bash
go run ./cmd/tduex
```

対話モードでは以下を順に聞きます。

1. 取得単位
2. year
3. term
4. day
5. period
6. export 形式

`USER_ID` / `PASSWORD` が未設定なら、その前に入力を求めて `.setting` に保存します。

### **CLI 実行**

_classes only_

```bash
tduex classes -year 2025 -term 1 -format json,csv,xlsx
```

_full export_

```bash
tduex full -year 2025 -term 1 -day 2 -period 1 -format json,csv,xlsx,ics
```

### **Help**

```bash
tduex --help
```

## **Install**

### **最短**

```bash
go install ./cmd/tduex
```

これが一番楽です。`GOBIN` または `$(go env GOPATH)/bin` に `tduex` が入ります。

### **Clone 後にインストール**

```bash
git clone <repo>
cd TDUScheExport
sh scripts/install.sh
```

権限が不要な場所を自動で選んでインストールします。  
`/usr/local/bin` が書けない場合は `~/.local/bin` などに入ります。

### **Makefile**

```bash
make install-user
```

システム全体に入れたい場合だけ:

```bash
make install
```

### **ローカルビルド**

```bash
make build
./tduex
```

## **Export Formats**

### **classes**

- `json`
- `csv`
- `xlsx`

### **full**

- `json`
- `csv`
- `xlsx`
- `ics`

`ics` は event の日時が解釈できるものだけを書き出します。

## **Save Dialog**

- macOS
  Finder 系の保存ダイアログを表示
- Windows
  Explorer 系の保存ダイアログを表示

保存ダイアログを使わず `out/` 配下に自動保存したい場合は `-dialog=false` を使います。

```bash
tduex full -year 2025 -term 1 -format json,csv -dialog=false
```

## **Environment Variables**

主に必要なものは以下です。

- `ALLOW_DOMAIN`
- `BASE_URL`
- `LOGIN_URL`

認証情報は `.setting` に保存できます。

- `USER_ID`
- `PASSWORD`

例:

```env
ALLOW_DOMAIN=els.sa.dendai.ac.jp
BASE_URL=https://els.sa.dendai.ac.jp
LOGIN_URL=https://els.sa.dendai.ac.jp/webclass/login.php
```

## **Technical Highlights**

- scraping / parser / service / appconfig で責務を分離しています
- `classes` と `full` の取得粒度を分けています
- 複数形式への export を同じ取得結果からまとめて行えます
- 資格情報が無ければ初回実行時に入力して `.setting` に保存できます
- 保存ダイアログを使う GUI 形式と、CLI 的な自動保存の両方に対応しています
- ログはコンソール出力のみです

## **License**

MIT
