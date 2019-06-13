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
  "github.com/disintegration/imaging"
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
  f, err := os.Create("failed.txt")
  if err != nil {
    panic(err)
  }

  w := bufio.NewWriter(f)
  //Main loop
OUTER:
  for _, document := range Documents {
    fmt.Printf("S8Guid: %v\r\n", document.S8Guid)
    //pdf := gofpdf.New("L", "pt", "Letter", "")
    startPriGuid := document.PriGuids[0]
    existingImageFile, err := os.Open(startPriGuid+".png")
    defer existingImageFile.Close();
    if (err != nil) {
      log.Println("Can't open "+startPriGuid+".png - Aborting "+document.S8Guid)
      continue OUTER
    }
    curImage, _, err := image.DecodeConfig(existingImageFile)
    if err != nil {
      log.Println(startPriGuid+".png - Aborting "+document.S8Guid, err)
      _, err := w.WriteString("Can't decode "+startPriGuid+".png - Aborting "+document.S8Guid+"\n")
      if err != nil {
        panic(err)
      }
      w.Flush()
      continue OUTER
    }
    pdf := gofpdf.NewCustom(&gofpdf.InitType {
      UnitStr:  "pt",
      Size: gofpdf.SizeType{Wd: float64(curImage.Width), Ht: float64(curImage.Height)},
      OrientationStr: "P",
    })
    pdf.SetFont("Arial", "B", 16)

    //pdf.AddPageFormat("L", gofpdf.SizeType{Wd: float64(curImage.Width), Ht: float64(curImage.Height)})
    for idx, priGuid := range document.PriGuids {
      //if idx > 0 {
        existingImageFile, err := os.Open(priGuid+".png")
        defer existingImageFile.Close();
        if (err != nil) {
          log.Println("Can't open "+priGuid+".png - Aborting "+document.S8Guid)
          _, err := w.WriteString("Can't open "+priGuid+".png - Aborting "+document.S8Guid+"\n")
          if err != nil {
            panic(err)
          }
          w.Flush()
          continue OUTER
        }
        curImage, _, err := image.DecodeConfig(existingImageFile)
        if err != nil {
          log.Println(priGuid+".png - Aborting "+document.S8Guid, err)
          _, err := w.WriteString("Can't Decode "+priGuid+".png - Aborting "+document.S8Guid+"\n")
          if err != nil {
            panic(err)
          }
          w.Flush()
          continue OUTER
        }

        src, err := imaging.Open(priGuid+".png")
        if err != nil {
          log.Println("failed to open image "+ priGuid +": %v", err)
          log.Println("Aborting "+document.S8Guid)
          _, err := w.WriteString("failed to open "+priGuid+".png - Aborting "+document.S8Guid+"\n")
          if err != nil {
            panic(err)
          }
          w.Flush()
          continue OUTER
        }
        if (curImage.Width > curImage.Height) {
          src = imaging.Rotate90(src)
        }
        err = imaging.Save(src, priGuid+".jpg")
        if err != nil {
          log.Println("failed to save image: %v", err)
          _, err := w.WriteString("failed to save image "+priGuid+".jpg - Aborting "+document.S8Guid+"\n")
          if err != nil {
            panic(err)
          }
          w.Flush()
          continue OUTER
        }

        pdf.AddPageFormat("P", gofpdf.SizeType{Wd: float64(curImage.Width), Ht: float64(curImage.Height)})
      //}
      //pdf.SetX(0)
      wd, ht, u := pdf.PageSize(idx)
      log.Println("", idx, wd, u, ht, u)
      fmt.Printf("%v\r\n",priGuid)
      var opt gofpdf.ImageOptions
      opt.ImageType = "jpg"
      pdf.ImageOptions(priGuid+".jpg",0,0,float64(curImage.Width),float64(curImage.Height),false,opt, 0, "")
    }
    pdf.OutputFileAndClose(document.S8Guid+".pdf")
  }
}
