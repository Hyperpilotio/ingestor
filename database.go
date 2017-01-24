package main

import (
	"strings"

	"gopkg.in/mgo.v2"

	"github.com/golang/glog"
	"github.com/hyperpilotio/ingestor/capturer"
	"github.com/spf13/viper"
)

type DB struct {
	DbType       string
	Url          string
	DatabaseName string
	TableName    string
	DBSession    *mgo.Session
}

// Connect to the database
func ConnectDB(config *viper.Viper) (*DB, error) {
	dbConfig := config.Sub("database")
	dbType := strings.ToLower(dbConfig.GetString("type"))
	dbUrl := dbConfig.GetString("url")
	dbName := dbConfig.GetString("databaseName")
	tbName := dbConfig.GetString("tableName")
	dbClient := &DB{
		DbType:       dbType,
		Url:          dbUrl,
		DatabaseName: dbName,
		TableName:    tbName,
	}

	if dbType == "mongo" {
		err := dbClient.openMongo()
		if err != nil {
			return nil, err
		}
	}

	return dbClient, nil
}

func (db *DB) openMongo() error {
	sess, err := mgo.Dial(db.Url)
	if err != nil {
		glog.Errorln("Couldn't connect to the MongoDB")
		return err
	}

	db.DBSession = sess

	return nil
}

func (db *DB) Insert(deployments capturer.Deployments) {
	var data []interface{}
	session := db.DBSession.Copy()
	defer session.Close()

	if db.DbType == "mongo" {
		c := session.DB(db.DatabaseName).C(db.TableName)

		data = append(data, deployments)
		err := c.Insert(data...)
		if err != nil {
			glog.Errorln(err)
		}
	}
}

// Disconnect from the database
func (db *DB) Close() {
	if db.DbType == "mongo" {
		db.DBSession.Close()
	}
}
