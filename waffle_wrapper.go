package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	// ローカルのエンジンを実行するならこう
	// cmd := exec.Command("/Users/mwatanabe/path/to/engine")
	// cmd := exec.Command("C:\\Users\\mwata\\wafffle_yane\\HoneyWaffle.exe")
	cmd := exec.Command(
		"ssh",
		"-i",
		"ec2-key-file.pem",
		"ubuntu@ec2-123-45-67-89.ap-northeast-1.compute.amazonaws.com",
		"/home/ubuntu/YourEngineName")

	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer cmdStdin.Close()

	// cmd.Stdin = os.Stdin
	scanner := bufio.NewScanner(os.Stdin)
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	quitCh := make(chan bool, 1)
	go func() {
		cmd.Wait()
		quitCh <- true
	}()

	go func() {
		<-quitCh
		os.Exit(0)
	}()

	isBlack := true
	isNoColor := true

	for scanner.Scan() {
		text := scanner.Text()
		if isNoColor {
			if strings.HasPrefix(text, "position") {
				// 先手なら"position startpos"だけ
				tmpArr := strings.Split(text, " ")
				if len(tmpArr) == 2 {
					isBlack = true
				} else {
					isBlack = false
				}
				isNoColor = false
			}
		}
		// 必要なら持ち時間を改ざんする
		if strings.HasPrefix(text, "go") {
			tmpArr := strings.Split(text, " ")
			// "go btime 318000 wtime 312000 binc 10000 winc 10000"
			if tmpArr[1] != "ponder" {
				// ponderの時にもやるのは怖すぎる
				btime, _ := strconv.Atoi(tmpArr[2])
				wtime, _ := strconv.Atoi(tmpArr[4])
				if isBlack {
					// 相手の持ち時間が1分を切っていたら、時間攻め
					if (btime > wtime) && (60000 > wtime) {
						btime = 5000
					}
					// 持ち時間が離されないように。
					if (wtime - btime) > 180000 {
						btime = 10000
					}
				} else {
					// 相手の持ち時間が1分を切っていたら、時間攻め
					if (wtime > btime) && (60000 > btime) {
						wtime = 5000
					}
					// 持ち時間が離されないように。
					if (btime - wtime) > 180000 {
						wtime = 10000
					}
				}
				text = tmpArr[0] + " btime " + strconv.Itoa(btime) + " wtime " + strconv.Itoa(wtime) + " binc " + tmpArr[6] + " winc " + tmpArr[8]
			}
		}
		io.WriteString(cmdStdin, text+"\n")
	}
}
