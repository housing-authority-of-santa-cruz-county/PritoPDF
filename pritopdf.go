package main
import (
  "bufio"
  "encoding/csv"
  "fmt"
  "io"
  "log"
  "os"
  "time"
  "strconv"
  "github.com/jung-kurt/gofpdf"
  "image"
)
const (
  layoutPri = "2006-01-02 15:04:05.000"
)
type PriRecord struct {
    PriGuid string
    S8Guid  string
    ArchiveTime time.Time
    Collection  string
    ChildOrder  int64
}

type Document struct {
  S8Guid  string
  PriGuids  []string
}

func main() {
  csvFile, _ := os.Open("guids.csv")
  reader := csv.NewReader(bufio.NewReader(csvFile))
  var PriRecords []PriRecord
  var Documents []Document
  confMap := make(map[string]int)

  for {
    line, error := reader.Read()
    if error == io.EOF {
      break
    } else if error != nil {
      log.Fatal(error)
    }
    archiveTime, err := time.Parse(layoutPri,line[2])
    if err != nil {
      log.Println("input is")
      log.Println(line[2])
      log.Fatal(err)
    }

    childOrder, err := strconv.ParseInt (line[4],0,0)
    if err != nil {
      log.Fatal(err)
    }
    PriRecords = append(PriRecords, PriRecord{
      PriGuid: line[0],
      S8Guid:   line[1],
      ArchiveTime: archiveTime,
      Collection: line[3],
      ChildOrder: childOrder,
    })
    s8Guid := line[1]
    if idx, ok := confMap[s8Guid]; ok {
      Documents[idx].PriGuids = append(Documents[idx].PriGuids, line[0])
    } else {
      confMap[s8Guid] = len(Documents)
      Documents = append(Documents, Document {
        S8Guid: line[1],
        PriGuids: []string{line[0]},
      })
    }

  }
  //fmt.Printf("%v",Documents[0])
  //Main loop
OUTER:
  for _, document := range Documents {
    fmt.Printf("S8Guid: %v\r\n", document.S8Guid)
    //pdf := gofpdf.New("P", "mm", "Letter", "")
    pdf := gofpdf.NewCustom(&gofpdf.InitType {
      UnitStr:  "pt",
      Size: gofpdf.SizeType{Wd: 2550, Ht: 3506},
    })
    pdf.SetFont("Arial", "B", 16)
    for _, priGuid := range document.PriGuids {
      pdf.SetX(0)
      fmt.Printf("%v\r\n",priGuid)
      var opt gofpdf.ImageOptions
      opt.ImageType = "png"
      existingImageFile, err := os.Open(priGuid+".png")
      defer existingImageFile.Close();
      if (err != nil) {
        log.Println("Can't open "+priGuid+".png - Aborting "+document.S8Guid)
        continue OUTER
      } else {
        image, _, err := image.DecodeConfig(existingImageFile)
        if err != nil {
          log.Println(priGuid+".png - Aborting "+document.S8Guid, err)
          continue OUTER
        }
        log.Println(image.Height)
        pdf.ImageOptions(priGuid+".png",0,0,2550,3506,true,opt, 0, "")
      }
      //pdf.AddPage()
    }
    pdf.OutputFileAndClose(document.S8Guid+".pdf")
  }

/*  pdf := gofpdf.New("P", "mm", "A4", "")
  pdf.AddPage()
  pdf.SetFont("Arial", "B", 16)
  pdf.Cell(40, 10, "Hello, world")
  pdf.OutputFileAndClose("hello.pdf")
*/
}
