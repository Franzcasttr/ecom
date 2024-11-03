package main

import (
	"database/sql"
	"log"

	"github.com/Franzcasttr/ecom/cmd/api"
	env "github.com/Franzcasttr/ecom/config"
	"github.com/Franzcasttr/ecom/db"
	"github.com/go-sql-driver/mysql"
)

func main(){
	db, err := db.NewSQLStorage(mysql.Config{
		User: env.Envs.DBUser,
		Passwd: env.Envs.DBPassword,
		Addr:env.Envs.DBAddress,
		DBName:env.Envs.DBName,
		Net:"tcp",
		AllowNativePasswords:true,
		ParseTime:true,
	})

	if err != nil {
		log.Fatal(err)
	}
	initStorage(db)
	
	server := api.NewAPIServer(":8000", db);
	if err := server.Run(); err != nil{
		log.Fatal(err);
	}

}

func initStorage(db *sql.DB){
	err:= db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: successfully connected")
}