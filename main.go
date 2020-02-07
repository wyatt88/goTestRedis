package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
	
	"github.com/go-redis/redis/v7"
)

func getValue(addr string,db int,key string,taskId int,ch chan string) {
	client := redis.NewClient(&redis.Options{
		Addr:               addr,
		Dialer:             nil,
		OnConnect:          nil,
		Password:           "",
		DB:                 db,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        0,
		ReadTimeout:        0,
		WriteTimeout:       0,
		PoolSize:           1,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
		Limiter:            nil,
	})
	startRedisTime := time.Now()
	val, err := client.Get(key).Result()
	result := val
	if err != nil {
		result = err.Error()
	}
	endRedisTime := time.Now()
	if len(val) > 20 {
		result = "-"
	}
	ch <- fmt.Sprintf("task id %d, the duration is %d ms, result is %s.\n",taskId,endRedisTime.Sub(startRedisTime).Nanoseconds()/1000/1000,result)
	return
}

func getValueUsePool(client *redis.Client,key string,taskId int,ch chan string) {
	startRedisTime := time.Now()
	val, err := client.Get(key).Result()
	result := val
	if err != nil {
		result = err.Error()
	}
	endRedisTime := time.Now()
	if len(val) > 20 {
		result = "-"
	}
	ch <- fmt.Sprintf("task id %d, the duration is %d ms, result is %s.\n",taskId,endRedisTime.Sub(startRedisTime).Nanoseconds()/1000/1000,result)
	return
}

func main()  {
	var  redisAddr string
	var  redisPort int
	var  redisDB int
	var  keyName string
	var threadCount int
	var  usePool int
	flag.StringVar(&redisAddr,"addr","127.0.0.1","The address of Redis")
	flag.IntVar(&redisPort,"port",6379,"The port number of Redis")
	flag.IntVar(&redisDB,"db",0,"use DB, Default is 0")
	flag.StringVar(&keyName,"key","name","The name of key. Default is foo")
	flag.IntVar(&threadCount,"threads",200,"The number of threads.")
	flag.IntVar(&usePool,"pool",50,"The poolsize,if poolsize is 0,don't use pool")
	flag.Parse()
	chs := make([]chan string, threadCount)
	fullAddr := fmt.Sprintf("%s:%d",redisAddr,redisPort)
	startTime := time.Now()
	fmt.Printf("MultiRun start, Runtime CPU: %d\n",runtime.NumCPU())
	if usePool == 0 {
		for i := 0; i < threadCount ; i++  {
			chs[i] = make(chan string)
			go getValue(fullAddr,redisDB,keyName,i,chs[i])
		}
	} else if usePool > 0  {
		client := redis.NewClient(&redis.Options{
			Addr:               fullAddr,
			Dialer:             nil,
			OnConnect:          nil,
			Password:           "",
			DB:                 redisDB,
			MaxRetries:         0,
			MinRetryBackoff:    0,
			MaxRetryBackoff:    0,
			DialTimeout:        0,
			ReadTimeout:        0,
			WriteTimeout:       0,
			PoolSize:           usePool,
			MinIdleConns:       0,
			MaxConnAge:         0,
			PoolTimeout:        0,
			IdleTimeout:        0,
			IdleCheckFrequency: 0,
			TLSConfig:          nil,
			Limiter:            nil,
		})
		for i := 0; i < threadCount ; i++  {
			chs[i] = make(chan string)
			go getValueUsePool(client,keyName,i,chs[i])
		}
	}
	
	for _, ch := range chs {
		fmt.Println(<-ch)
	}
	endTime := time.Now()
	fmt.Printf("MultiRun finished. Process time %d. Number of tasks is %d.\n", endTime.Sub(startTime).Nanoseconds()/1000/1000, threadCount)
}
