package control

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"mania/constant"
	"mania/logger"
	"strconv"
	"time"
)

type Storage struct {
	MongoDB   *mongo.Client
	RedisConn *redis.Pool
}

var Store Storage

const Online = 1
const Offline = 0

func (s *Storage) Connect(configSrvName string) error {
	defer func(start time.Time) {
		fmt.Println("Connect:", time.Now().Sub(start).Milliseconds())
	}(time.Now())

	mongoPool := GetMongodbSource()
	mongodb, err := ConnectMongoDB(mongoPool, configSrvName)
	if err != nil {
		return err
	}

	redisPool := GetRedisSource()
	var cache *redis.Pool
	if len(redisPool.SentinelPath) > 0 {
		//cache, err = InitRedisSentinelConnPool(redisPool)
	} else {
		cache, err = ConnectRedis(redisPool)
	}
	if err != nil {
		return err
	}

	s.MongoDB = mongodb
	s.RedisConn = cache

	return nil
}

func (s *Storage) CloseDB() {
	_ = s.DelUserOnline(constant.ServerName)

	err := s.RedisConn.Close()
	if err != nil {
		logger.Error("Redis关闭异常", "err", err.Error())
	}
	err = s.MongoDB.Disconnect(context.TODO())
	if err != nil {
		logger.Error("MongoDB关闭异常", "err", err.Error())
	}

	logger.Error("release DB success")
}

func (s *Storage) DelUserOnline(node string) (err error) {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Warn("DelUserOnline", "err", err.Error())
		}
	}(conn)

	key := fmt.Sprintf("%s_USERONLINE_BIT", node)
	_, err = conn.Do("DEL", key)
	if err != nil {
		logger.Error("DelUserOnline", "err", err.Error())
		return
	}

	return
}

func (s *Storage) UserOnline(uid int, node string) error {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("UserOnline", "err", err.Error())
		}
	}(conn)
	key := fmt.Sprintf("%s_USERONLINE_BIT", node)
	_, err := conn.Do("SETBIT", key, uid, Online)
	if err != nil {
		return err
	}

	_, err = conn.Do("HSET", "USERONLINE", uid, node)
	if err != nil {
		return err
	}

	logger.Info("", "key", "UserOnline", "uid", uid, "node", node, "redis_key", key)
	return nil
}
func (s *Storage) UserOffline(uid int, node string) error {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("UserOffline", "err", err.Error())
		}
	}(conn)
	key := fmt.Sprintf("%s_USERONLINE_BIT", node)
	_, err := conn.Do("SETBIT", key, uid, Offline)
	if err != nil {
		return err
	}

	_, err = conn.Do("HDEL", "USERONLINE", uid, node)
	if err != nil {
		return err
	}

	logger.Info("", "key", "UserOffline", "uid", uid, "node", node, "redis_key", key)
	return nil
}
func (s *Storage) IsUserOnline(uid int) bool {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("IsUserOnline", "err", err.Error())
		}
	}(conn)

	key := fmt.Sprintf("%s_USERONLINE_BIT", constant.ServerName)
	if ok, _ := redis.Bool(conn.Do("GETBIT", key, uid)); ok {
		return true
	}
	return false
}

func (s *Storage) concurrentUser(node string) (num int) {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("concurrentUser", "err", err.Error())
		}
	}(conn)
	key := fmt.Sprintf("%s_USERONLINE_BIT", node)
	_, err := redis.Int(conn.Do("BITCOUNT", key))
	if err != nil {
		logger.Error("concurrentUser", "err", err.Error())
		return
	}
	return
}

// AllUserOnline 获取所有在线用户
func (s *Storage) AllUserOnline() (users []int) {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("AllUserOnline", "err", err.Error())
		}
	}(conn)

	keys, err := redis.Strings(conn.Do("HKEYS", "USERONLINE"))
	if err != nil {
		logger.Error("AllUserOnline", "err", err.Error())
		return
	}

	for _, key := range keys {
		uid, _ := strconv.Atoi(key)
		users = append(users, uid)
	}

	return
}

func (s *Storage) CreateUserID() (int, error) {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("CreateUserID", "err", err.Error())
		}
	}(conn)
	uid, err := redis.Int(conn.Do("INCR", "USERID"))
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func (s *Storage) GetUser(filter bson.D) (singleResult *mongo.SingleResult, err error) {
	collection := s.MongoDB.Database(constant.DB).Collection("user")

	res := collection.FindOne(context.TODO(), filter)

	return res, res.Err()
}

func (s *Storage) CreateUser(document interface{}) error {
	collection := s.MongoDB.Database(constant.DB).Collection("user")

	_, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetUserToken(filter bson.D) (resp *mongo.SingleResult, err error) {
	collection := s.MongoDB.Database(constant.DB).Collection("user_token")

	res := collection.FindOne(context.TODO(), filter)

	return res, res.Err()
}

func (s *Storage) CreateUserToken(document interface{}) error {
	collection := s.MongoDB.Database(constant.DB).Collection("user_token")

	_, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) UpdateUserToken(filter bson.D, update bson.D) error {
	collection := s.MongoDB.Database(constant.DB).Collection("user_token")
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}
func (s *Storage) UpdateUser(filter bson.D, update bson.D) error {
	collection := s.MongoDB.Database(constant.DB).Collection("user")
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

// CreateMapId 新增地图编号
func (s *Storage) CreateMapId() (int, error) {
	conn := s.RedisConn.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("CreateMapId", "err", err.Error())
		}
	}(conn)
	roomId, err := redis.Int(conn.Do("incr", "MapId"))
	if err != nil {
		return roomId, err
	}

	return roomId, nil
}

// CreateMap 创建地图
func (s *Storage) CreateMap(document interface{}) error {
	collection := s.MongoDB.Database(constant.DB).Collection("map")

	_, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		return err
	}
	return nil
}
