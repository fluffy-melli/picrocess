# Picrocess - Go 이미지 처리 라이브러리

`picrocess`는 Go 언어로 이미지 처리 기능을 제공하는 라이브러리입니다. 아래는 주요 기능과 사용법 예제입니다.

## 설치

다음 명령어를 사용하여 `picrocess` 모듈을 다운로드할 수 있습니다.

```go
$ go get github.com/fluffy-melli/picrocess
```

## 코드 예제

### 1. 기본 이미지 생성

#### 1.1. 500x700 크기의 투명 이미지 만들기

```go
base := picrocess.NewImage(500, 700, picrocess.NewRGBA(0, 0, 0, 0))
```

#### 1.2. 300x300 크기의 빨간 반투명 이미지 만들기

```go
red := picrocess.NewImage(300, 300, picrocess.NewRGBA(255, 0, 0, 105))
```

### 2. 이미지 로드

#### 2.1. 파일에서 이미지 불러오기

```go
image, err := picrocess.LoadImage("./~.png")
if err != nil {
    panic(err)
}
```

#### 2.2. URL 에서 이미지 불러오기

```go
image, err := picrocess.ImageURL("https://~")
if err != nil {
    panic(err)
}
```

#### 2.3. 폰트 파일 로드

```go
font, err := picrocess.LoadFont("./NanumGothic.ttf")
if err != nil {
    panic(err)
}
```

### 3. 이미지 조작

#### 3.1. 이미지 크기 변경

이미지의 크기를 200x200으로 축소합니다.

```go
image.Resize(200, 200)
```

#### 3.2. 다른 이미지를 투명 이미지에 추가

빨간 반투명 이미지를 100, 200 좌표에 추가합니다.

```go
base.Overlay(red, picrocess.NewOffset(100, 200))
```

#### 3.3. 불러온 이미지를 투명 이미지에 추가

불러온 이미지를 0, 0 좌표에 추가합니다.

```go
base.Overlay(image, picrocess.NewOffset(0, 0))
```

#### 3.4. 텍스트 추가

폰트를 사용하여 노란색 텍스트 `~`를 10, 150 좌표에 50크기로 추가합니다.

```go
base.DrawString(font, picrocess.NewRGBA(255, 255, 0), picrocess.NewOffset(10, 150), 50, "~")
```

### 4. 이미지 저장

#### 4.1. 이미지 저장

이미지를 PNG 형식으로 저장합니다.

```go
base.SaveAsPNG("./index.png")
```

#### 4.2. 이미지 자르기

투명 이미지의 0, 0 좌표에서 500, 500까지 잘라서 새로운 이미지로 저장합니다.

```go
cor := base.Crop(picrocess.NewRect(0, 0, 500, 500))
cor.SaveAsPNG("./index2.png")
```

### 5. 이미지 `[]byte`로 변환

이미지를 `[]byte`로 변환하여 저장할 수도 있습니다. `ToPNG`와 `ToJPG` 메서드를 사용하면 이미지를 각각 PNG와 JPG 형식으로 변환할 수 있습니다.

#### 5.1. 이미지 PNG 형식으로 `[]byte`로 변환

```go
pngData, err := base.ToPNG()
if err != nil {
    panic(err)
}
// pngData는 []byte 형식의 PNG 이미지 데이터입니다.
```

#### 5.2. 이미지 JPG 형식으로 `[]byte`로 변환

```go
jpgData, err := base.ToJPG(90) // 90은 JPG 품질(0-100)입니다.
if err != nil {
    panic(err)
}
// jpgData는 []byte 형식의 JPG 이미지 데이터입니다.
```


## 라이센스

이 라이브러리는 MIT 라이센스를 따릅니다. 자세한 내용은 `LICENSE` 파일을 참고하세요.
