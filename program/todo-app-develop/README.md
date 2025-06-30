## 概要
- 非常に簡単な todo-app
- `index.html.erb`, `show.html.erb`, `edit.html.erb` はAIを使用して作成

## 実行手順
1. クローン
```
git clone https://github.com/Yu-7777/todo-app.git
```
2. ビルド
```
docker compose build
```
3. DB用意
```
docker compose run web rails db:parepare
```
4. 実行
```
dokcer compose up
```
5. アクセス
- localhost:3000 にて確認可能
