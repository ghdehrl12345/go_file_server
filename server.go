package main

import (
	"fmt"
	"io"
	"net/http" // HTTP 관련 기능(서버, 요청 처리 등)을 다루는 패키지입니다.
	"os"
	"path/filepath"
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

		ext := filepath.Ext(filename)
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

		fmt.Fprintf(w, "파일 업로드 성공! 파일명: %s (MIME: %s)", handler.Filename, mimeType)
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
			<h2> Go 파일 업로드 서비스</h2>
			<p> JPG 또는 PNG 파일만 업로드할 수 있습니다. </p>

			<form action="/" method="POST" enctype="multipart/form-data">
			
				<label for="file">파일 선택:</label>
				<input type="file" id="file" name="uploadFile">
				<br><br>
				<input type="submit" value="업로드 시작">
			</form>
		`)
	}
}

func main() {
	// http.HandleFunc은 "만약 / 주소로 요청이 오면, 'handler' 함수를 실행해!"라고 등록하는 것입니다.
	http.HandleFunc("/", handler)

	fs := http.FileServer(http.Dir("uploads"))

	http.Handle("/files/", http.StripPrefix("/files/", fs))

	fmt.Println("서버가 8080 포트에서 실행 중")

	os.MkdirAll("uploads", os.ModePerm)
	// http.ListenAndServe는 실제로 8080 포트에서 서버를 시작하고 요청을 '듣기' 시작합니다.
	// 만약 서버 시작에 실패하면(예: 다른 프로그램이 8080 포트를 이미 사용 중) 에러를 반환합니다.
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("서버 시작 실패:", err)
	}
}
