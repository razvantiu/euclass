package main

import "log"

func main() {
	err := genkey()

	if err != nil {
		log.Fatalln(err)
	}
}

func genkey() error {

	return nil

}
