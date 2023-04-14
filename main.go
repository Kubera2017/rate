package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
rate_limiter_handler производит:
1. определение маски подсети запроса
2. проверку наличия подсети в списке заблокированных
  - в случае блокировки ответ с ошибкой
  - в случае истечения времени блока - разблокировка и успешный ответ

3. увеличение счетчика для подсети
  - в случае новой подсети - создание нового счетчика

4. проверка частоты запросов на условия блока
  - в случае разницы во времени между визитами больше cooldown - сброс счетчика
  - в случае превышения частоты - блокировка

Ипользуются структуры:
LockMap - информация о заблокированных подсетях и времени наложения блока
SubnetCountMap - информация о частоте запросов от подсетей
*/
func rate_limiter_handler(
	locks *LockMap,
	subnet_count_map SubnetCountMap,
) func(w http.ResponseWriter, req *http.Request) {
	max_requests := 3
	var period int64 = (10 * time.Second).Nanoseconds()

	return func(w http.ResponseWriter, r *http.Request) {

		ip := r.Header["X-Forwarded-For"][0]
		subnet := strings.Split(ip, ".")[0:3]
		var subnet_hash [3]uint8
		for i, col := range subnet {
			col_int, _ := strconv.Atoi(col)
			subnet_hash[i] = uint8(col_int)
		}
		// [127 0 0]

		log.Println("Mask", subnet_hash, "lock_map", locks.subnet_map)

		locked, unlocked := locks.check(subnet_hash, period)
		if locked {
			log.Println("Mask", subnet_hash, "Locked err")
			http.Error(w, fmt.Sprint(locked), http.StatusTooManyRequests)
		} else {
			log.Println("Mask", subnet_hash, "subnet_count_map", subnet_count_map)
			subnet_count, has := subnet_count_map[subnet_hash]
			if unlocked {
				log.Println("unblock reset")
				subnet_count.reset()
				log.Println("unblocked reset", subnet_count)
			}
			if !has {
				subnet_count = new_SubnetCount()
				log.Println("new", subnet_count)
			}

			if !unlocked && has && subnet_count.increment_and_check(max_requests, period) {
				locks.block(subnet_hash)
			}
			subnet_count_map[subnet_hash] = subnet_count

			fmt.Fprintf(w, "hello\n")
		}

	}
}

func main() {
	lock_map, subnet_count_map := get_blank_maps()

	http.HandleFunc("/", rate_limiter_handler(lock_map, subnet_count_map))

	http.ListenAndServe(":3000", nil)
}
