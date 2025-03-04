// 创建应用数据库用户
db.createUser({
    user: "fiber_web",
    pwd: "fiber_web_password",
    roles: [
        {
            role: "readWrite",
            db: "fiber_web"
        }
    ]
}); 