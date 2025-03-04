#!/bin/bash
set -e

# 使用root用户连接MongoDB
mongosh <<EOF
use admin
db.auth('$MONGO_INITDB_ROOT_USERNAME', '$MONGO_INITDB_ROOT_PASSWORD')

// 创建应用数据库
use $MONGO_INITDB_DATABASE

// 创建应用专用用户
db.createUser({
  user: 'fiber_web',
  pwd: 'fiber_web_password',
  roles: [
    {
      role: 'readWrite',
      db: '$MONGO_INITDB_DATABASE'
    }
  ]
})
EOF 