package dao

import (
    "fmt"
    "errors"
    "sona/broker/logic"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

type ConfigureDocument struct {
    ServiceKey string `bson:"serviceKey"`
    Version uint `bson:"version"`
    ConfKeys []string `bson:"confKeys"`
    ConfValues []string `bson:"confValues"`
}

//获取collection以便操作
func getCollection() (*mgo.Session, *mgo.Collection) {
    url := fmt.Sprintf("%s/%d", logic.GConf.DbHost, logic.GConf.DbPort)
    session, err := mgo.Dial(url)
    if err != nil {
        fmt.Printf("%s\n", err)
        return nil, nil
    }

    collection := session.DB(logic.GConf.DbName).C(logic.GConf.DbCollectionName)
    if collection == nil {
        session.Close()
        fmt.Printf("no database %s or collection %s\n", logic.GConf.DbName, logic.GConf.DbCollectionName)
        return nil, nil
    }
    return session, collection
}

//加载所有数据
func ReloadAllData() ([]*ConfigureDocument, error) {
    session, collection := getCollection()
    if collection == nil {
        return nil, errors.New("database error")
    }
    defer session.Close()
    results := make([]ConfigureDocument, 0)
    err := collection.Find(bson.M{}).All(&results)
    if err != nil {
        return nil, err
    }

    data := make([]*ConfigureDocument, 0)
    for _, result := range results {
        data = append(data, &result)
    }
    return data, nil
}

//新增数据，发起自admin添加
func AddDocument(serviceKey string, version uint, confKeys []string, confValues []string) error {
    session, collection := getCollection()
    if collection == nil {
        return errors.New("database error")
    }
    defer session.Close()
    return collection.Insert(&ConfigureDocument{
        ServiceKey: serviceKey,
        Version: version,
        ConfKeys: confKeys,
        ConfValues: confValues,
    })
}

//修改数据
func UpdateDocument(serviceKey string, version uint, confKeys []string, confValues []string) error {
    session, collection := getCollection()
    if collection == nil {
        return errors.New("database error")
    }
    defer session.Close()
    return collection.Update(bson.M{"serviceKey":serviceKey},
    &ConfigureDocument{
        ServiceKey: serviceKey,
        Version: version,
        ConfKeys: confKeys,
        ConfValues: confValues,
    })
}