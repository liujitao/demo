 package common
 
 // 导入用户数据
 if count, _ := db.Collection("user").CountDocuments(ctx, bson.M{}); count == 0 {
    content, err := ioutil.ReadFile("common/userData.json")
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }

    var users []user.UserModel
    err = json.Unmarshal(content, &users)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }

    var document interface{}
    for _, u := range users {
        document = bson.D{
            {"_id", primitive.NewObjectID()},
            {"id", u.ID},
            {"user_name", u.UserName},
            {"mobile", u.Mobile},
            {"email", u.Email},
            {"password", common.SetPassword(u.Password)},
            {"create_at", time.Now()},
            {"update_at", time.Now()},
        }
        _, err = db.Collection("user").InsertOne(ctx, document)
        if err != nil {
            log.Fatal("Error insert user data: ", err)
        }
    }
}

// 导入团队数据
if count, _ := db.Collection("team").CountDocuments(ctx, bson.M{}); count == 0 {
    content, err := ioutil.ReadFile("common/teamData.json")
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }

    var teams []team.TeamModel
    err = json.Unmarshal(content, &teams)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }

    var document interface{}
    for _, t := range teams {
        document = bson.D{
            {"_id", primitive.NewObjectID()},
            {"id", t.ID},
            {"team_name", t.TeamName},
            {"describe", t.Describe},
            {"user_uuid", t.User},
            {"create_at", time.Now()},
            {"update_at", time.Now()},
        }
        _, err = db.Collection("team").InsertOne(ctx, document)
        if err != nil {
            log.Fatal("Error insert team data: ", err)
        }
    }
}

// 导入角色数据
if count, _ := db.Collection("role").CountDocuments(ctx, bson.M{}); count == 0 {
    content, err := ioutil.ReadFile("common/roleData.json")
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }

    var roles []role.RoleModel
    err = json.Unmarshal(content, &roles)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }

    var document interface{}
    for _, r := range roles {
        document = bson.D{
            {"_id", primitive.NewObjectID()},
            {"id", r.ID},
            {"role_name", r.RoleName},
            {"describe", r.Describe},
            {"user_uuid", r.User},
            {"create_at", time.Now()},
            {"update_at", time.Now()},
        }
        _, err = db.Collection("role").InsertOne(ctx, document)
        if err != nil {
            log.Fatal("Error insert role data: ", err)
        }
    }
}