package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"b612.me/starlog"
	"github.com/spf13/cobra"
)

var emo = NewEmojis()
var regOld, regNew, regFilter string
var allowCate []string
var useJson bool

func init() {
	cmdRoot.AddCommand(cmdMain, cmdShow)
	cmdMain.Flags().StringVarP(&emo.SaveFolders, "savepath", "s", "./Emojis", "Path of the emoji download folder. (Default is 'Emojis' in the current path)")
	cmdMain.Flags().StringVarP(&emo.AuthCookie, "cookie", "c", "", "Value of '_session_id' for the authentication cookie. (Optional)")
	cmdMain.Flags().BoolVarP(&emo.IgnoreErr, "ignore-error", "i", false, "Ignore download errors (Default is false)")
	cmdMain.Flags().BoolVarP(&emo.Zip2Tarfile, "zip", "z", true, "Compress into a tar.gz file (Default is true)")
	cmdMain.Flags().BoolVarP(&emo.DeletedOriginIfZip, "delete-after-zip", "d", false, "Delete the folder after compressing (Default is false)")
	cmdMain.Flags().StringSliceVarP(&allowCate, "allow-download-category", "a", []string{}, "Categories to download (Optional)")
	cmdMain.Flags().StringVarP(&regFilter, "filter", "f", "", "RegEx for the emoji name whitelist. (Optional)")
	cmdMain.Flags().StringVarP(&regOld, "replace-old", "o", "", "RegEx for replacing old emoji names with new ones. (Optional)")
	cmdMain.Flags().StringVarP(&regNew, "replace-new", "r", "", "RegEx for replacing emoji names with a new string. (Optional)")
	cmdMain.Flags().IntVarP(&emo.Threads, "threads", "n", 16, "Number of concurrent downloads (Optional, default is 16)")
	cmdMain.Flags().StringVarP(&emo.Proxy, "proxy", "p", "", "Use a proxy (Optional)")
	cmdRoot.PersistentFlags().BoolVarP(&useJson, "use-json-file", "j", false, "Use a JSON file (Optional, default is false)")
}

var cmdRoot = &cobra.Command{
	Use:   "get",
	Short: "Mastodon Emoji Downloader",
}

var cmdMain = &cobra.Command{
	Use:   "get",
	Short: "Mastodon Emoji Downloader",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			starlog.Errorln("Please enter the URL for the Mastodon server or the path for the JSON file.")
			os.Exit(1)
		}
		if regFilter != "" {
			rpgF, err := regexp.Compile(regFilter)
			if err != nil {
				starlog.Errorln("Invalid RegEx: ", regFilter, err)
			}
			emo.filterRp = rpgF
		}
		if regOld != "" {
			rpgF, err := regexp.Compile(regOld)
			if err != nil {
				starlog.Errorln("Invalid RegEx: ", regOld, err)
			}
			emo.rpCodeOld = rpgF
			emo.rpNew = regNew
		}
		if emo.Threads <= 0 {
			starlog.Errorln("Concurrent downloads cannot be less than or equal to 0. ", emo.Threads)
			os.Exit(3)
		}
		url := strings.ToLower(strings.TrimSpace(args[0]))
		if !useJson && strings.Index(url, "https://") != 0 {
			url = "https://" + url
		}
		err := emo.LoadAndParse(url, !useJson)
		if err != nil {
			starlog.Errorln("Load emoji lists failed: ", err)
			os.Exit(4)
		}
		if len(allowCate) != 0 {
			myMap := emo.CategoryCount()
			for _, v := range allowCate {
				if _, ok := myMap[v]; !ok {
					starlog.Errorln(v, "The category doesn't exist. Please check your input.")
					os.Exit(5)
				}
				emo.AllowCategories[v] = true
			}
		}
		err = emo.Download(showFn)
		if err != nil {
			starlog.Errorln("Download failed. ", err)
		}
		starlog.Infoln("Done!")
	},
}

var cmdShow = &cobra.Command{
	Use:   "category",
	Args:  cobra.ExactArgs(1),
	Short: "Show Mastodon Emoji Categories",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			starlog.Errorln("Please enter the URL for the Mastodon server or the path for the JSON file.")
			os.Exit(1)
		}
		url := strings.ToLower(strings.TrimSpace(args[0]))
		if !useJson && strings.Index(url, "https://") != 0 {
			url = "https://" + url
		}
		err := emo.LoadAndParse(url, !useJson)
		if err != nil {
			starlog.Errorln("Load emoji lists failed: ", err)
			os.Exit(4)
		}
		ct, orderSlice, allCat := emo.Counts()
		fmt.Println("Results for each category: ")
		fmt.Printf("%-5s %-10s %-28s\n", "No.", "Emojis", "Name")
		for k, v := range orderSlice {
			fmt.Printf("%-5v %-10d %-28s\n", k+1, allCat[v], v)
		}
		starlog.Green("%d emojis found in %d categories\n", len(orderSlice), ct)
		starlog.Infoln("Done! Saved to: ", emo.SaveFolders)
	},
}
