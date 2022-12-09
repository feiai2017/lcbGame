package model

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"mania/constant"
	"mania/logger"
	"mania/tcpx"
	"strconv"
	"sync"
)

var users = make(map[int]*User)
var userMute = new(sync.RWMutex)

type User struct {
	Conn        *tcpx.Context `bson:"-"`
	Package     string        `bson:"-"`
	UID         int
	SnakeLadder *SnakeLadder
}

func Get(uid int) (*User, error) {
	userMute.RLock()
	usr, ok := users[uid]
	userMute.RUnlock()
	if !ok {
		err := fmt.Sprintf("usr %d not exist", uid)
		return nil, errors.New(err)
	}
	return usr, nil
}
func Set(usr *User) error {
	if usr == nil || usr.UID == 0 {
		return errors.New("usr is nil")
	}
	userMute.Lock()
	users[usr.UID] = usr
	userMute.Unlock()
	return nil
}
func Destroy(uid int) {
	userMute.Lock()
	delete(users, uid)
	userMute.Unlock()
}
func (usr *User) Log(log *logger.Logger) {
	log.Append("",
		"uid", usr.UID)
}

type Token struct {
	Duid       string
	UID        int
	Credential string
}

func (t *Token) GenerateCertificate() {
	raw := strconv.Itoa(t.UID) + t.Duid
	l := strconv.Itoa(len(raw))
	sum := sha256.Sum256([]byte(raw + l + constant.ServerName))
	sumsha1 := sha1.Sum(sum[:])
	t.Credential = hex.EncodeToString(sumsha1[:])

	return
}
