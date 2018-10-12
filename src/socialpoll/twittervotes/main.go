package main

import (
	"github.com/bitly/go-nsq"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var stopLock sync.Mutex

	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)

	go func() {
		//what is this for?
		<-signalChan
		stopLock.Lock()
		stop = true
		stopLock.Unlock()
		log.Println("stopping")
		stopChan <- struct{}{}
		closeConn()
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if e := dialdb(); e != nil {
		log.Fatalln(`failed to dial mongodb`, e)
	}
	defer closedb()

	//start things

	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	stream := startTwitterStream(stopChan, votes)

	//do
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stopLock.Lock()

			if stop {
				stopLock.Unlock()
				return
			}
			stopLock.Unlock()
		}
	}()
	//
	<-stream
	close(votes)
	<-publisherStoppedChan

}

var db *mgo.Session

func dialdb() error {
	var err error
	host := `kvn1`
	log.Println("dialing mongodb: "+host)
	db, err = mgo.Dial(host)
	return err
}
func closedb() {
	db.Close()
	log.Println("closed database connection")
}

type poll struct {
	Options []string
}

func loadOptions() ([]string, error) {
	var options []string
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	var p poll
	for iter.Next(&p) {

		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

type tweet struct {
	Text string
}

func publishVotes(votes <-chan string) <-chan struct{} {

	stopchan := make(chan struct{}, 1)

	producer, _ := nsq.NewProducer("kvn1:4150", nsq.NewConfig())

	go func() {
		for vote := range votes {
			producer.Publish("votes", []byte(vote))
		}
		log.Println("publisher stopping")
		producer.Stop()
		log.Println("producer stopped")
		stopchan <- struct{}{}
	}()

	return stopchan
}
