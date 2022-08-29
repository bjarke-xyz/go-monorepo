package application

import (
	"ShoppingList-Backend/internal/pkg/item"
	"ShoppingList-Backend/internal/pkg/list"
	"ShoppingList-Backend/pkg/config"
	"ShoppingList-Backend/pkg/db"
	"ShoppingList-Backend/pkg/server"
	"fmt"

	"github.com/gomodule/redigo/redis"
	socketio "github.com/googollee/go-socket.io"
)

type Application struct {
	Cfg         *config.Config
	Queries     *Repositories
	Controllers *Controllers
	Redis       *redis.Pool
	Srv         *server.Server
	SocketIo    *socketio.Server
}

func Get(cfg *config.Config) (*Application, error) {
	db, err := db.Get(cfg.GetDBConnStr())
	if err != nil {
		return nil, err
	}

	repos := &Repositories{
		Item: &item.ItemRepository{
			DB: db.Client,
		},
		List: &list.ListRepository{
			DB: db.Client,
		},
	}

	controllers := &Controllers{
		Item: item.NewItemController(repos.Item),
		List: list.NewListController(repos.Item, repos.List),
	}

	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.GetRedisConnStr(), redis.DialClientName(cfg.GetRedisClientName()), redis.DialUsername(cfg.GetRedisUser()), redis.DialPassword(cfg.GetRedisPassword()))
		},
	}

	socketServer := socketio.NewServer(nil)
	_, err = socketServer.Adapter(&socketio.RedisAdapterOptions{
		Network:  "tcp",
		Addr:     cfg.GetRedisConnStr(),
		Prefix:   cfg.GetRedisPrefix() + ".socket.io",
		Password: cfg.GetRedisPassword(),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create redis socket io adapter: %w", err)
	}

	return &Application{
		Cfg:         cfg,
		Queries:     repos,
		Redis:       redisPool,
		Controllers: controllers,
		SocketIo:    socketServer,
	}, nil
}
