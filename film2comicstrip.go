package main

import (
	"fmt" //
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	//    "path/filepath"
	"bufio" // http://rosettacode.org/mw/index.php?title=Read_a_file_line_by_line&action=edit&section=33
	"flag"
	"strconv"
	"strings"
)

// cha	"code.google.com/p/go-charset/charset"
// 		r, _ := cha.NewReader("windows-1251", bytes.NewReader(dec)) // convert from CP932 to UTF-8

var (
	version      = 2.04
	copyright    = "(c) 2015 transient"
	framespersec = flag.Float64("fps", 25.0, "Frames per second of film") //float32 = 25.0
	filmname     = flag.String("film", "", "Film name")
	srtfilename  = flag.String("srtfile", "", "Srt file name")
	workDir      = flag.String("wrkdir", ".", "Work dir")
	//	enc     = flag.String("enc", "cp1251", "Input Encoding")
	step 		= flag.Int("step", 1, "Stage of work (1-5); 0 means all stages")
	vdub 		= flag.String("vdub", "", "Abs path to VirtualDub, like c:\\programs\\virtualdub.exe")
	ofmt		= flag.String("of", "djvu", "Output format: djvu or cbz")
)

func isWhiteSp(r rune) bool {
	if r == '\u0020' || r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' || r == '\u00A0' {
		return (true)
	} else {
		return (false)
	}
}

func isBlankLine(s string) bool {
	for _, char := range s {
		if !isWhiteSp(char) {
			return (false)
		}
	}
	return (true)
}

// format: 01:30:55,760
func timestr2float(s string) float64 {
	stampMs := strings.Split(s, ",")
	stampHMS := strings.Split(stampMs[0], ":")
	hs, _ := strconv.Atoi(stampHMS[0])
	ms, _ := strconv.Atoi(stampHMS[1])
	ss, _ := strconv.Atoi(stampHMS[2])
	sumsecs := 3600*hs + 60*ms + ss
	msec, _ := strconv.Atoi(stampMs[1])
	secFloat := float64(sumsecs) + (float64(msec) / 1000)
	return (secFloat)
}

func sec2frame(x float64, fps float64) int {
	return (int(x * fps))
}

func stringGuard(s string) string {
	return (strings.Replace(s, "\"", "'", -1))
}

func main() {

	log.SetFlags(log.Lshortfile)

	flag.Parse()

	fmt.Println("2015 by transient. ver.", version)

	if *filmname == "" {
		log.Fatal("film name must be defined!")
	}
	if *srtfilename == "" {
		log.Fatal("srt file name must be defined!")
	}

	var pwd string // current and start dir
	//	var pwd,prevd string // current and start dir
	var err error

	if *workDir == "." {
		pwd, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		//	    prevd, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		//newd := "C:\\TMP";
		err = os.Chdir(*workDir)
		if err != nil {
			log.Fatal(err)
		}
		pwd, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	/////////////////////////////////////////////////////////////////////
	/// 1st stage
	/////////////////////////////////////////////////////////////////////

	if *step == 1 {

		//  srtfilename := "test.srt"
		inputFile, err := os.Open(*srtfilename) // открываем текущий srt-файл, получаем ссылку

		if err != nil {
			log.Fatal("Error opening input file:", err)
		}
		// Closes the file when we leave the scope of the current function,
		// this makes sure we never forget to close the file if the
		// function can exit in multiple places.
		defer inputFile.Close()

		avsTrimfilename := "trimmer.avs"
		avsTrimfile, err := os.Create(avsTrimfilename)
		if err != nil {
			log.Fatal(err)
		}
		wTrimFile := bufio.NewWriter(avsTrimfile)
		defer avsTrimfile.Close()

		avsSubtfilename := "subtitles.avs"
		avsSubtfile, err := os.Create(avsSubtfilename)
		if err != nil {
			log.Fatal(err)
		}
		wSubtFile := bufio.NewWriter(avsSubtfile)
		defer avsSubtfile.Close()

		firststring := true
		scanner := bufio.NewScanner(inputFile)

		//  var numtitrstr string
		var txt string
		//  var outputstr string
		var framenum string

		// scanner.Scan() advances to the next token returning false if an error was encountered
		for scanner.Scan() { // читаем построчно srt-файл
			str := scanner.Text()

			// надо обрезать BOM в самой первой строке -- из-за этого возможны ошибки с определением пустой строки и комментария
			if firststring {
				str = strings.TrimPrefix(str, "\xEF\xBB\xBF")
				firststring = false
			}

			if isBlankLine(str) {
				// , align=2, y=400, size=30, first_frame=2715, last_frame=2715, lsp=2
				fmt.Fprintln(wSubtFile, "Subtitle(align=2,y=400,size=30,lsp=2,\""+txt+"\",first_frame="+framenum+",last_frame="+framenum+")")
				txt = ""

				continue
			} // это пустая строка

			//      if (str == "13") {fmt.Println("13!!!")}
			//      numtitr,err := strconv.Atoi(str) // получаем номер субтитра
			_, err := strconv.Atoi(str) // если номер нам не нужен

			//      if (err != nil) { continue } // это не номер был ?? ну или выполнять другое действие
			if err == nil {
				// fmt.Println(numtitr, corrLenOfTitr(numtitr));
				//        numtitrstr = corrLenOfTitr(numtitr)
				continue
			} // это не номер был ?? ну или выполнять другое действие

			timestamp := strings.Split(str, " --> ")
			if len(timestamp) == 2 {
				//fmt.Println(timestamp[0], timestamp[1]);
				t1 := timestr2float(timestamp[0])
				t2 := timestr2float(timestamp[1])
				//fmt.Println(t1,t2);

				arith_aver := (t1 + t2) / 2
				framenum = fmt.Sprint(sec2frame(arith_aver, *framespersec))

				_, err = wTrimFile.WriteString("+Trim(" + framenum + "," + framenum + ")")
				if err != nil {
					log.Fatal(err)
				}
				continue
			}

			if txt == "" {
				txt = stringGuard(str)
			} else {
				txt += "\\n" + stringGuard(str)
			}

		}

		wTrimFile.Flush()
		wSubtFile.Flush()

	}

	/////////////////////////////////////////////////////////////////////
	/// 2nd stage
	/////////////////////////////////////////////////////////////////////

	if *step == 2 {

		avsJobfname := "job.avs"
		avsJobfile, err := os.Create(avsJobfname)
		if err != nil {
			log.Fatal(err)
		}
		defer avsJobfile.Close()

		_, err = avsJobfile.WriteString("AVISource(\"" + *filmname + "\",false)\n")
		if err != nil {
			log.Fatal(err)
		}

		_, err = avsJobfile.WriteString("#DirectShowSource(\"" + *filmname + "\",fps=25.000,audio=false)\n#BilinearResize(720,400)\n")
		if err != nil {
			log.Fatal(err)
		}

		inSubtitles, err := os.Open("subtitles.avs") // открываем avs-файл выбранных подписей, получаем ссылку
		if err != nil {
			log.Fatal("Error opening subtitles.avs:", err)
		}
		defer inSubtitles.Close()

		scanner := bufio.NewScanner(inSubtitles)

		for scanner.Scan() { // читаем построчно srt-файл
			str := scanner.Text()
			avsJobfile.WriteString("\n" + str)
		}

		/*
			inTrims, err := os.Open("trimmer.avs") // открываем avs-файл выбранных подписей, получаем ссылку
			if err != nil {
				log.Fatal("Error opening trimmer.avs:", err)
			}
			defer inTrims.Close()
		*/

		rawBytes, err := ioutil.ReadFile("trimmer.avs")
		if err != nil {
			//			return nil, err
			log.Fatal("Error reading trimmer.avs:", err)
		}
		trimstr := string(rawBytes)

		strings.TrimPrefix(trimstr, "+")

		avsJobfile.WriteString("\n\n" + trimstr)

		if jdir, err := os.Stat("jpeg"); os.IsNotExist(err) {
			err := os.Mkdir("jpeg", os.ModeDir)
			if err != nil {
				log.Fatal("Error creating \"jpeg\" dir:", err)
			}
		} else {
			if !jdir.IsDir() {
				log.Fatal("\"jpeg\" is file, not a dir:")
			}
		}

		vcfJobfname := "job.vcf"
		vcfJobfile, err := os.Create(vcfJobfname)
		if err != nil {
			log.Fatal(err)
		}
		defer vcfJobfile.Close()

//		fmt.Println("see pwd=",pwd)
		_, err = vcfJobfile.WriteString("VirtualDub.Open(U\"job.avs\",\"\",0);\nVirtualDub.audio.SetSource(0);\n")
		_, err = vcfJobfile.WriteString("VirtualDub.SaveImageSequence(U\"" + pwd + "\\jpeg\\\", \".jpg\", 3, 2, 95);\n")
		_, err = vcfJobfile.WriteString("VirtualDub.Close();")
		if err != nil {
			log.Fatal(err)
		}
	}

	/////////////////////////////////////////////////////////////////////
	/// 3rd stage
	/////////////////////////////////////////////////////////////////////

	if *step == 3 {
		if *vdub == "" {
			*vdub = "VirtualDub.exe"
		}
		//		VirtualDub.exe /s job.vcf /x
		vDCmd := exec.Command(*vdub, "/s", "job.vcf", "/x")
		_, err := vDCmd.Output()
		if err != nil {
			log.Fatal(err)
		}

	}

	/////////////////////////////////////////////////////////////////////
	/// 4th stage
	/////////////////////////////////////////////////////////////////////

	if *step == 4 {
		if *vdub == "" {
			*vdub = "VirtualDub.exe"
		}
		//		VirtualDub.exe /s job.vcf /x
		vDCmd := exec.Command(*vdub, "/s", "job.vcf", "/x")
		_, err := vDCmd.Output()
		if err != nil {
			log.Fatal(err)
		}

	}

}
