package hashing

import (
	"strconv"
	"time"
)

var chars = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "0", "9", "8", "7", "6", "5", "4", "3", "2", "1", "/", ".", ":", "-", " "}
var charsRandom = []string{"H", "G", "$", "I", "S", "z", "O", "e", "7", "U", "M", "Y", "r", "l", "K", "m", "a", "v", "+", "C", "y", "6", "J", "-", "i", "4", "1", "F", "t", "B", "2", "p", "8", "h", "A", "g", "c", "X", "T", "u", "o", "k", "W", "V", "w", "N", "s", "P", "D", "b", "0", "x", "L", "Q", "@", "9", "Z", "3", "j", "n", "q", "f", "!", "E", "R", "5", "d"}
var charsMap = map[string]string{"a": "H", "b": "G", "c": "$", "d": "I", "e": "S", "f": "z", "g": "O", "h": "e", "i": "7", "j": "U", "k": "M", "l": "Y", "m": "r", "n": "l", "o": "K", "p": "m", "q": "a", "r": "v", "s": "+", "t": "C", "u": "y", "v": "6", "w": "J", "x": "-", "y": "i", "z": "4", "A": "1", "B": "F", "C": "t", "D": "B", "E": "2", "F": "p", "G": "8", "H": "h", "I": "A", "J": "g", "K": "c", "L": "X", "M": "T", "N": "u", "O": "o", "P": "k", "Q": "W", "R": "V", "S": "w", "T": "N", "U": "s", "V": "P", "W": "D", "X": "b", "Y": "0", "Z": "x", "0": "L", "9": "Q", "8": "@", "7": "9", "6": "Z", "5": "3", "4": "j", "3": "n", "2": "q", "1": "f", "/": "!", ".": "E", ":": "R", "-": "5", " ": "d"}
var charsMap2 = map[string]string{"I": "d", "v": "r", "1": "A", "o": "O", "3": "5", "r": "m", "l": "n", "t": "C", "f": "1", "7": "i", "C": "t", "h": "H", "A": "I", "k": "P", "0": "Y", "L": "0", "!": "/", "e": "h", "6": "v", "i": "y", "D": "W", "b": "X", "Q": "9", "Z": "6", "j": "4", "2": "E", "c": "K", "X": "L", "u": "N", "y": "u", "J": "w", "4": "z", "5": "-", "U": "j", "K": "o", "p": "F", "q": "2", "Y": "l", "F": "B", "g": "J", "n": "3", "a": "q", "s": "U", "R": ":", "S": "e", "B": "D", "V": "R", "@": "8", "-": "x", "W": "Q", "x": "Z", "G": "b", "M": "k", "T": "M", "z": "f", "m": "p", "8": "G", "N": "T", "d": " ", "H": "a", "$": "c", "+": "s", "w": "S", "P": "V", "E": ".", "O": "g", "9": "7"}

func GenerateHash(link string) string {
	if len(link) > 6 {
		if "/media" == link[:6] {
			link = link[6:]
		} else if "media" == link[:5] {
			link = link[5:]
		}
	}

	heshLink := ""
	heshH := ""
	heshD := ""
	date := time.Now().UTC().Add(15 * time.Minute)
	h := date.Format("15:04:05")
	d := date.Format("02.01.2006")
	for _, v := range h {
		heshH += charsMap[string(v)]
	}
	i := 0
	newMap := make(map[string]string, 0)
	for _, v := range charsRandom[date.Second():] {
		newMap[chars[i]] = v
		i++
	}
	for _, v := range charsRandom[:date.Second()] {
		newMap[chars[i]] = v
		i++
	}
	for _, v := range d {
		heshD += newMap[string(v)]
	}
	for _, v := range link {
		if val, ok := newMap[string(v)]; ok {
			heshLink += val
		} else {
			heshLink += string(v)
		}
	}
	return "/media/" + heshLink + newMap[" "] + heshD + newMap[" "] + heshH
}

func OpenHash(hash string) string {
	if len(hash) > 10 {
		hashH := hash[len(hash)-8:]
		link := ""
		h := ""
		for _, v := range hashH {
			h += charsMap2[string(v)]
		}
		min, _ := strconv.Atoi(h[6:])
		newMap := make(map[string]string, 0)
		i := 0
		for _, v := range charsRandom[min:] {
			newMap[v] = chars[i]
			i++
		}
		for _, v := range charsRandom[:min] {
			newMap[v] = chars[i]
			i++
		}
		for _, v := range hash[:len(hash)-8] {
			if val, ok := newMap[string(v)]; ok {
				link += val
			} else {
				link += string(v)
			}
		}
		return link + h
	}
	return ""
}
