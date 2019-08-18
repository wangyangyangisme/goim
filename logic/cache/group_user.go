package cache

import (
	"errors"
	"goim/logic/db"
	"goim/logic/model"
	"goim/public/imctx"
	"goim/public/logger"
	"strconv"
	"time"

	"github.com/json-iterator/go"
)

const (
	groupUserKey = "group_user:"
	groupUserExp = 2 * time.Hour
)

var ErrResult = errors.New("error redis result")

type groupUserCache struct{}

var GroupUserCache = new(groupUserCache)

func (*groupUserCache) Key(appId, groupId int64) string {
	return groupUserKey + strconv.FormatInt(appId, 10) + strconv.FormatInt(groupId, 10)
}

// 保存群组所有用户的信息
func (c *groupUserCache) SetAll(ctx *imctx.Context, appId, groupId int64, userInfos []model.GroupUser) error {
	users := make(map[string]interface{}, len(userInfos)+1)
	for _, userInfo := range userInfos {
		bytes, err := jsoniter.Marshal(userInfo)
		if err != nil {
			logger.Sugar.Error(err)
			return err
		}

		users[strconv.FormatInt(userInfo.UserId, 10)] = bytes
	}

	key := c.Key(appId, groupId)
	err := db.RedisCli.HMSet(key, users).Err()
	if err != nil {
		logger.Sugar.Error(err)
		return err
	}

	err = db.RedisCli.Expire(key, groupUserExp).Err()
	if err != nil {
		logger.Sugar.Error(err)
		return err
	}
	return nil
}

// Get 获取用户在群组中的信息
func (c *groupUserCache) Get(ctx *imctx.Context, appId, groupId, userId int64) (*model.GroupUser, error) {
	var groupUser model.GroupUser
	err := hget(c.Key(appId, groupId), strconv.FormatInt(userId, 10), &groupUser)
	if err != nil {
		logger.Sugar.Error(err)
		return nil, err
	}

	return &groupUser, nil
}

// Set 设置用户在群组中的信息
func (c *groupUserCache) Set(ctx *imctx.Context, appId, groupId, userId int64) (*model.GroupUser, error) {
	var groupUser model.GroupUser
	err := hset(c.Key(appId, groupId), strconv.FormatInt(userId, 10), &groupUser)
	if err != nil {
		logger.Sugar.Error(err)
		return nil, err
	}

	return &groupUser, nil
}

func (c *groupUserCache) Del(ctx *imctx.Context, appId, groupId, userId int64) error {
	err := hdel(c.Key(appId, groupId), strconv.FormatInt(userId, 10))
	if err != nil {
		logger.Sugar.Error(err)
		return err
	}

	return nil
}
