package control

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"mania/model"
	"time"
)

func ConnectMongoDB(dataSource *model.Mongodb, configSrvName string) (*mongo.Client, error) {
	// 设置客户端连接配置
	// add 20220607
	// write concern : majority , timeout: 1s
	// read concern : majority
	clientOptions := options.Client().ApplyURI(dataSource.Path).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(time.Second))).
		SetReadPreference(readpref.Primary()).
		SetReadConcern(readconcern.Majority())
	if dataSource.CaFilePath != "" {
		uri := fmt.Sprintf(dataSource.Path, dataSource.CaFilePath, dataSource.CertPemPath)
		credential := options.Credential{
			AuthMechanism: "MONGODB-X509",
		}
		clientOptions = options.Client().ApplyURI(uri).SetAuth(credential)
		clientOptions.TLSConfig.InsecureSkipVerify = true
	}

	// 连接到MongoDB
	db, err := mongo.Connect(context.TODO(), clientOptions,
		options.Client().SetMaxPoolSize(10000),
		options.Client().SetMinPoolSize(100),
		options.Client().SetAppName(configSrvName),
		options.Client().SetWriteConcern(writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(time.Second))),
		options.Client().SetReadPreference(readpref.Primary()),
		options.Client().SetReadConcern(readconcern.Majority()))
	if err != nil {
		err = errors.Wrap(err, "connect mongodb failed")
		return nil, err
	}

	// 检查连接
	err = db.Ping(context.TODO(), nil)
	if err != nil {
		err = errors.Wrap(err, "ping mongodb failed")
		return nil, err
	}
	return db, nil
}
