# mailAuto
급여명세서 자동 첨부 메일 전송


1. input 폴더에 급여명세서를 넣는다. PDF파일로 여러장 구성되어 있어야 함
2. "메일 자동" 폴더 안에 .env 파일을 작성하고 네이버 웍스 IMAP/SMTP 계정과 비밀번호 작성
3. main.go 를 컴파일하여 실행한다. ex) start.exe
4. 2번 실행시, input 폴더 안에 급여명세서를 spilt.exe가 나눠서 output폴더에 저장한다.
5. start.exe에서 output폴더에 나눠진 급여명세서에서 글씨를 추출해 비교 후 "급여작업_자동화.csv"에서 가져온 이름과 일치 시 연결된 이메일로 전송
