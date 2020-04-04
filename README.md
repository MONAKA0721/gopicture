# 開発環境
```
# イメージのビルド
$ docker-compose build

# バックグラウンドでアプリを立ち上げ
$ docker-compose up -d

# チェック用にコンテナの中に入る
$ docker run --rm -it gopicture-image /bin/sh
```
ブラウザでアクセス  
http://localhost:8888


# デプロイ
```
staging に自動デプロイ
$ git push origin develop

production に自動デプロイ
$ git push origin master
```
staging  
https://gopicture-docker-stg.herokuapp.com  
production  
https://gopicture-docker.herokuapp.com
