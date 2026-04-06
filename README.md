# **tduex**

東京電機大学の webclass から講義一覧やイベント情報を取得して、  
複数形式で export する CLI ツールです。

## **Overview**

- webclss の時間割や講義ページを手作業で確認する手間を減らすために作成しました。
- 講義一覧だけを取るモードと、event まで辿って取得するモードを切り替えられます。
- export 形式は `JSON` / `CSV` / `XLSX` / `ICS` に対応しています。
- macOS では Finder 系、Windows では Explorer 系の保存ダイアログから保存先を選べます。

## **Architecture**

### **Tech Stack**

- Backend: Go
- Scraping: Colly / goquery
- Export: JSON / CSV / XLSX / ICS
- Config: `.env` / `~/.config/tduex/.setting` / `~/.config/tduex/.usersetting`
- Install: GitHub Releases / PowerShell / Makefile / shell script

## **Project Structure**

```text
.
├── cmd
│   └── tduex
├── internal
│   ├── appconfig
│   ├── logger
│   ├── parser
│   ├── scraping
│   ├── service
│   └── tduexcli
├── scripts
│   ├── install.ps1
│   └── install.sh
├── .github
│   └── workflows
│       └── release.yml
├── Makefile
├── go.mod
└── README.md
```

## **Main Flow**

`tduex` が行う基本処理は以下です。

1. 設定ファイルを読み込む
2. 必要なら認証情報を入力して `~/.config/tduex/.usersetting` に保存する
3. webclassから講義一覧または event 情報を取得する
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

`USER_ID` / `PASSWORD` が未設定なら、その前に入力を求めて `~/.config/tduex/.usersetting` に保存します。  
fetch に失敗した場合は、`.usersetting` の `USER_ID` / `PASSWORD` を確認するようメッセージを表示します。

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

### **Windows**

PowerShell だけでインストールできます。`Go` も `sh` も不要です。

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
```

このスクリプトは GitHub Releases から `tduex.exe` を取得して、`%LOCALAPPDATA%\Programs\tduex\bin` に入れ、必要ならユーザー PATH に追加します。

まだ Release を作っていない開発中の状態では、リポジトリ直下の `tduex.exe` か `dist\tduex.exe` も使えます。`Go` が入っていれば最後にローカルビルドへフォールバックします。

特定バージョンを入れたい場合:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1 -Version v0.1.0
```

手元の exe を使いたい場合:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1 -SourceExe .\tduex.exe
```

### **macOS / Linux 最短**

```bash
go install github.com/med-000/tduex/cmd/tduex@latest
```

`GOBIN` または `$(go env GOPATH)/bin` に `tduex` が入ります。

### **Clone 後にインストール**

```bash
git clone <repo>
cd tduex
sh scripts/install.sh
```

権限が不要な場所を自動で選んでインストールします。  
`/usr/local/bin` が書けない場合は `~/.local/bin` などに入ります。

Windows では `sh` が前提になるので、この手順ではなく `scripts/install.ps1` を使ってください。

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

### **リリース配布**

`v*` タグを push すると [release.yml](/Users/med/Documents/App_dev/product/tduex/.github/workflows/release.yml) が各 OS 向けバイナリを作り、GitHub Release に添付します。Windows の `install.ps1` はその配布物を使います。

## **Export Formats**

### **classes**

- `json`
- `csv`
- `xlsx`

返る内容:

- `json`
  `externalId`, `year`, `term`, `classes[]`
- `json` の `classes[]`
  `externalId`, `day`, `period`, `title`
- `json` の例

```json
{
  "externalId": "2025_1",
  "year": 2025,
  "term": 1,
  "classes": [
    {
      "externalId": "xxxxxxxxxxxxxxxx",
      "day": 2,
      "period": 1,
      "title": "コンピュータ構成"
    }
  ]
}
```

- `csv`
  1 行 1 講義
- `csv` / `xlsx` の列
  `externalId`, `year`, `term`, `day`, `period`, `title`
- `xlsx`
  `classes` シートに講義一覧を表形式で出力

### **full**

- `json`
- `csv`
- `xlsx`
- `ics`

返る内容:

- `json`
  `externalId`, `year`, `term`, `classes[]`
- `json` の `classes[]`
  `externalId`, `day`, `period`, `title`, `events[]`
- `json` の `events[]`
  `externalId`, `name`, `category`, `date`, `groupName`
- `json` の例

```json
{
  "externalId": "2025_1",
  "year": 2025,
  "term": 1,
  "classes": [
    {
      "externalId": "xxxxxxxxxxxxx",
      "day": 2,
      "period": 1,
      "title": "コンピュータ構成",
      "events": [
        {
          "externalId": "xxxxxxxxxxxxxxx",
          "name": "期末考査対策用自習教材を置いておきます",
          "category": "資料",
          "date": "",
          "groupName": "期末考査対策用"
        }
      ]
    }
  ]
}
```

- `csv`
  1 行 1 event
- `csv` / `xlsx` の列
  `classExternalId`, `year`, `term`, `day`, `period`, `classTitle`, `eventExternalId`, `eventName`, `category`, `date`, `groupName`
- `xlsx`
  `events` シートに event 一覧を表形式で出力
- `ics`
  event の日時をカレンダーイベントとして出力

`ics` は event の `date` が日時として解釈できるものだけを書き出します。

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

例:

`~/.config/tduex/.setting`

```env
ALLOW_DOMAIN=els.sa.dendai.ac.jp
BASE_URL=https://els.sa.dendai.ac.jp
LOGIN_URL=https://els.sa.dendai.ac.jp/webclass/login.php
```

`~/.config/tduex/.usersetting`

```env
USER_ID=your_user_id
PASSWORD=your_password
```

一般設定は `~/.config/tduex/.setting`、認証情報は `~/.config/tduex/.usersetting` に保存できます。  
互換のため、カレントディレクトリの `.setting` / `.usersetting` もあれば読み込みます。

## **Technical Highlights**

- scraping / parser / service / appconfig で責務を分離しています
- `classes` と `full` の取得粒度を分けています
- 複数形式への export を同じ取得結果からまとめて行えます
- 資格情報が無ければ初回実行時に入力して `.usersetting` に保存できます
- 保存ダイアログを使う GUI 形式と、CLI 的な自動保存の両方に対応しています
- ログはコンソール出力のみです

## **License**

MIT
