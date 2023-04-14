package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func get_blank_maps() (*LockMap, SubnetCountMap) {
	lock_map := LockMap{
		lock:       sync.RWMutex{},
		subnet_map: make(map[[3]uint8]int64),
	}

	subnet_count_map := make(map[[3]uint8]SubnetCount)

	return &lock_map, subnet_count_map
}

/*
LockMap структура состоит из map подсетей b времени наложения блока и глобального mutex на запись для этой map
*/
type LockMap struct {
	lock       sync.RWMutex
	subnet_map map[[3]uint8]int64
}

func (locks *LockMap) block(subnet [3]uint8) {
	log.Println("Block")
	locks.lock.RLock()
	locks.subnet_map[subnet] = time.Now().UnixNano()
	log.Println(locks.subnet_map)
	locks.lock.RUnlock()
}

func (locks *LockMap) unblock(subnet [3]uint8) {
	log.Println("Unblock")
	locks.lock.RLock()
	delete(locks.subnet_map, subnet)
	log.Println(locks.subnet_map)
	locks.lock.RUnlock()
	fmt.Print("Unblock done\n")
}

// check - returns 'locked' y/n, 'unblocked' y/n
func (locks *LockMap) check(subnet [3]uint8, period int64) (bool, bool) {
	locks.lock.Lock()

	locked := false
	unlocked := false

	blocked_timestamp, has := locks.subnet_map[subnet]
	if has {
		now := time.Now().UnixNano()
		if (now - blocked_timestamp) > period {
			locks.lock.Unlock()
			locks.unblock(subnet)
			unlocked = true
		} else {
			locked = true
		}
	}

	if !unlocked {
		locks.lock.Unlock()
	}
	return locked, unlocked
}

type SubnetCountMap = map[[3]uint8]SubnetCount

type SubnetCount struct {
	lock          *sync.Mutex
	track_started int64
	request_count int
}

func new_SubnetCount() SubnetCount {
	new_counter := SubnetCount{
		lock: &sync.Mutex{},
	}
	new_counter.reset()
	return new_counter
}

func (subnet_cnt *SubnetCount) reset() {
	subnet_cnt.lock.Lock()
	log.Println("reset")
	now := time.Now().UnixNano()
	subnet_cnt.request_count = 1
	subnet_cnt.track_started = now
	subnet_cnt.lock.Unlock()
}

// SubnetCount increment_and_check увеличиват счетчик и возвращает должна быть подсеть заблокирована или нет
func (subnet_cnt *SubnetCount) increment_and_check(max_requests int, period int64) bool {
	subnet_cnt.lock.Lock()
	now := time.Now().UnixNano()

	to_block := false

	subnet_cnt.request_count++
	if subnet_cnt.request_count >= max_requests {
		if now-subnet_cnt.track_started < period {
			subnet_cnt.lock.Unlock()
			to_block = true
		} else {
			subnet_cnt.lock.Unlock()
			subnet_cnt.reset()
		}
	} else {
		subnet_cnt.lock.Unlock()
	}
	return to_block
}
