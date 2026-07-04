package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"b612.me/stario"
	"b612.me/starlog"
)

var showFn = func(v Emoji, finished bool, err error) {
	if err != nil {
		starlog.Errorf("Download error for emoji '%-10s' in category '%s': %v \n", v.Category, v.ShortCode, err)
		return
	}
	if !finished {
		starlog.Noticef("Downloading emoji '%-10s' in category '%s' \n", v.Category, v.ShortCode)
		return
	}
	starlog.Infof("Downloaded emoji '%-10s' in category '%s' \n", v.Category, v.ShortCode)
}

func PromotMode() {
	starlog.SetLevelColor(starlog.LvNotice, []starlog.Attr{starlog.FgHiYellow})
	emo := NewEmojis()
	checkFn := func(fn func(emo *Emojis) int) {
		if code := fn(emo); code != 0 {
			stario.StopUntil("An error occurred. Press any key to exit.", "", true)
			os.Exit(code)
		}
	}
	for _, v := range []func(*Emojis) int{plParseJson, plGetDownloadChoice, plRegexp, plDownload} {
		checkFn(v)
	}
	stario.StopUntil("Done! Press any key to exit.", "", true)
}

func plParseJson(emo *Emojis) int {
	var fromSite bool
	var fpath string
	fmt.Println("Mastodon Emoji Downloader")
	fmt.Println("You are currently in interactive mode. If you want to use manual mode, run '--help' to view the usage instructions.")
	fmt.Println("Choose how to download:")
	fmt.Println("1. Download Mastodon emojis from a server URL")
	fmt.Println("2. Download Mastodon emojis from a JSON file")
	for {
		choice := stario.MessageBox("Enter: ", "0").MustInt()
		if choice != 1 && choice != 2 {
			starlog.Red("Please enter 1 or 2. Your input is %d\n", choice)
			continue
		}
		if choice == 1 {
			fromSite = true
		}
		break
	}
	for {
		if fromSite {
			fmt.Print("Enter the Mastodon server domain name, without 'https': ")
		} else {
			fmt.Print("Enter the path of the JSON file: ")
		}
		fpath = stario.MessageBox("", "").MustString()
		if fpath == "" {
			starlog.Red("Please enter the actual input. ")
			continue
		}
		break
	}
	if fromSite {
		if strings.Index(fpath, "https://") != 0 {
			fpath = "https://" + fpath
		}
		if stario.YesNo("Use a proxy? (y/N)", false) {
			emo.Proxy = stario.MessageBox("Enter the proxy address: ", "").MustString()
		}
		if stario.YesNo("Use a Mastodon cookie? (y/N)", false) {
			emo.AuthCookie = stario.MessageBox("Enter the value of '_session_id' for the authentication cookie: ", "").MustString()
		}
	}
	starlog.Infoln("Parsing...")
	err := emo.LoadAndParse(fpath, fromSite)
	if err != nil {
		starlog.Errorln("Parsing failed! Please check your input. ", err)
		return 1
	}
	return 0
}

func plGetDownloadChoice(emo *Emojis) int {
	ct, orderSlice, allCat := emo.Counts()
	fmt.Println("Results for each category: ")
	fmt.Printf("%-5s %-10s %-28s\n", "No.", "Emojis", "Name")
	for k, v := range orderSlice {
		fmt.Printf("%-5v %-10d %-28s\n", k+1, allCat[v], v)
	}
	starlog.Green("%d emojis found in %d categories\n", len(orderSlice), ct)
exitfor:
	for {
		choice, err := stario.MessageBox("Enter the category # you want to download. To download multiple categories, separate the numbers with commas. (Optional) ", "0").SliceInt(",")
		if err != nil {
			starlog.Errorln("Error: ", err, "Please check your input, or just press Enter.")
			continue
		}
		for _, v := range choice {
			if v == 0 {
				emo.AllowCategories = make(map[string]bool)
				fmt.Println("Will download every category.")
				break exitfor
			}
			emo.AllowCategories[orderSlice[v-1]] = true
			fmt.Println("Will download: ", orderSlice[v-1])
		}
		break
	}
	return 0
}

func plRegexp(emo *Emojis) int {
	if stario.YesNo("Enable the emoji name whitelist, using RegEx? (y/N): ", false) {
		for {
			rgpR, err := regexp.Compile(stario.MessageBox("Please enter the RegEx for the whitelist: ", "").MustString())
			if err != nil {
				starlog.Errorln("Invalid RegEx: ", err)
				continue
			}
			emo.filterRp = rgpR
			break
		}
	}
	if stario.YesNo("Enable the emoji name replacement, using RegEx? (y/N): ", false) {
		for {
			rgpR, err := regexp.Compile(stario.MessageBox("Enter the RegEx for replacing old emoji names with new ones: ", "").MustString())
			if err != nil {
				starlog.Errorln("Invalid RegEx: ", err)
				continue
			}
			emo.rpCodeOld = rgpR
			break
		}
		emo.rpNew = stario.MessageBox("Enter the new name to be replaced: ", "").MustString()
	}
	return 0
}

func plDownload(emo *Emojis) int {
	emo.SaveFolders = stario.MessageBox("Enter the path of the emoji download folder. (Default is 'Emojis' in the current path)： ", "./Emojis").MustString()
	emo.IgnoreErr = stario.YesNo("Ignore download errors? (Y/n): ", true)
	emo.Zip2Tarfile = stario.YesNo("Compress each category into a tar.gz file? (Y/n): ", true)
	if emo.Zip2Tarfile {
		emo.DeletedOriginIfZip = stario.YesNo("Delete the folder(s) after compressing? (y/N): ", false)
	}
	for {
		emo.Threads = stario.MessageBox("Number of concurrent downloads (Optional, default is 16)：", "16").MustInt()
		if emo.Threads <= 0 {
			starlog.Red("Invalid input: ", emo.Threads)
			continue
		}
		break
	}
	var fn func(v Emoji, finished bool, err error) = nil
	if stario.YesNo("Display the download log? (Y/n): ", true) {
		fn = showFn
	}
	starlog.Infoln("Downloading...")
	err := emo.Download(fn)
	if err != nil {
		starlog.Errorln("Download failed: ", err)
		return 1
	}
	starlog.Infoln("Emoji saved to: ", emo.SaveFolders)
	return 0
}
