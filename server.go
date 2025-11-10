package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 'handler' 함수는 누군가 웹사이트에 접속할 때 실행됩니다.
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Fprintf(w, "POST 요청, 파일 처리 시작")
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "폼 파싱 실패", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("uploadFile")
		if err != nil {
			fmt.Println("파일 가져오기 실패:", err)
			http.Error(w, "파일을 가져올 수 없습니다.", http.StatusBadRequest)
			return
		}

		defer file.Close()

		filename := filepath.Base(handler.Filename)

		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".png" {
			http.Error(w, "허영되지 않은 파일 확장자입니다 (.jpg, .png만 가능)", http.StatusBadRequest)
			return
		}

		fmt.Printf("정리된 파일명: %s\n, 확장자: %s\n", filename, ext)

		buffer := make([]byte, 512)

		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			http.Error(w, "파일 읽기 실패", http.StatusInternalServerError)
			return
		}

		mimeType := http.DetectContentType(buffer)
		fmt.Printf("감지된 MIME 타입: %s\n", mimeType)

		if mimeType == "image/jpeg" || mimeType == "image/png" {
			fmt.Println("MIME 타입 검증 통과!")
		} else {
			http.Error(w, "허용되지 않는 MIME 타입입니다 (jpeg, png만 가능)", http.StatusBadRequest)
			return
		}

		filePath := "uploads/" + filename
		_, err = os.Stat(filePath)

		if err == nil {
			http.Error(w, "파일이 이미 존재합니다.", http.StatusBadRequest)
			return
		}

		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "파일 생성 실패", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = dst.Write(buffer)

		if err != nil {
			http.Error(w, "파일 저장 실패 (버퍼 쓰기)", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "파일 저장 실패 (본문 복사)", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		entries, err := os.ReadDir("uploads")
		if err != nil {
			fmt.Println("uploads 디렉터리 읽기 실패 : ", err)
		}

		var files []string
		for _, entry := range entries {
			isNotHidden := !strings.HasPrefix(entry.Name(), ".")
			isNotDirectory := !entry.IsDir()
			if isNotHidden && isNotDirectory {
				files = append(files, entry.Name())
			}
		}
		htmlTemplate := `
			<h2>Go 파일 업로드 서비스</h2>
			<form action="/" method="POST" enctype="multipart/form-data">
				<label for="file">파일 선택 (JPG, PNG):</label>
				<input type="file" id="file" name="uploadFile">
				<input type="submit" value="업로드 시작">
			</form>
			
			<hr>
			
			<h3>업로드된 파일 목록 ({{len .}} 개)</h3>
			<div style="display: flex; flex-wrap: wrap; gap: 10px;">
				{{range .}}
					<div style="border: 1px solid #ccc; padding: 5px; text-align: center;">
						<a href="/files/{{.}}" target="_blank" download>
							<img src="/files/{{.}}" alt="{{.}}" width="200" height="200" style="object-fit: cover;">
							<br>
							<small>{{.}}</small>
						</a>
					</div>
				{{end}}
			</div>
		`

		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, "템플릿 파싱 오류", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, files)
	}
}

func main() {
	http.HandleFunc("/", handler)

	fs := http.FileServer(http.Dir("uploads"))

	http.Handle("/files/", http.StripPrefix("/files/", fs))

	fmt.Println("서버가 8080 포트에서 실행 중")

	os.MkdirAll("uploads", os.ModePerm)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("서버 시작 실패:", err)
	}
}
