package database

import (
	"errors"
	"strings"

	"gopkg.in/mgo.v2"

	"github.com/spf13/viper"
)

// TODO: Synchronize access with mutex
type MongoDB struct {
	Url          string
	DatabaseName string
	TableName    string
	DBSession    *mgo.Session
}

// Connect to the database
func NewDB(config *viper.Viper) (*MongoDB, error) {
	dbType := strings.ToLower(config.GetString("database.type"))
	dbUrl := config.GetString("database.url")
	dbName := config.GetString("database.databaseName")
	tbName := config.GetString("database.tableName")

	switch dbType {
	case "mongo":
		return &MongoDB{
			Url:          dbUrl,
			DatabaseName: dbName,
			TableName:    tbName,
		}, nil
	default:
		return nil, errors.New("Unsupported database type: " + dbType)
	}
}

func (db MongoDB) connect() (*mgo.Session, error) {
	sess, err := mgo.Dial(db.Url)
	if err != nil {
		return nil, errors.New("Unable to connect to mongo: " + err.Error())
	}

	return sess, nil
}

func (db MongoDB) Insert(data interface{}) error {
	session, sessionErr := db.connect()
	if sessionErr != nil {
		return errors.New("Unable to connect mongo: " + sessionErr.Error())
	}

	defer session.Close()

	c := session.DB(db.DatabaseName).C(db.TableName)

	err := c.Insert(data)
	if err != nil {
		return errors.New("Unable to insert data: " + err.Error())
	}

	return nil
}

func (db MongoDB) Upsert(selector map[string]interface{}, data interface{}) error {
	session, sessionErr := db.connect()
	if sessionErr != nil {
		return errors.New("Unable to connect mongo: " + sessionErr.Error())
	}

	defer session.Close()

	c := session.DB(db.DatabaseName).C(db.TableName)

	_, err := c.Upsert(selector, data)
	if err != nil {
		return errors.New("Unable to upsert data: " + err.Error())
	}

	return nil
}
