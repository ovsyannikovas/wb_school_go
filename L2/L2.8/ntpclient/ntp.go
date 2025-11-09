package ntpclient

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

func GetCurrentTime() (time.Time, error) {
	return ntp.Time("pool.ntp.org")
}

func PrintCurrentTime() int {
	currentTime, err := GetCurrentTime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения времени: %v\n", err)
		return 1
	}

	fmt.Println(currentTime.Format(time.RFC3339))
	return 0
}
