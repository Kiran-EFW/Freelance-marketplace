package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	googleVisionEndpoint = "https://vision.googleapis.com/v1/images:annotate"
)

// GoogleVisionProvider defines the interface for image analysis operations.
type GoogleVisionProvider interface {
	DetectText(ctx context.Context, imageData []byte) (*OCRResult, error)
	DetectLabels(ctx context.Context, imageData []byte) ([]Label, error)
	DetectFaces(ctx context.Context, imageData []byte) ([]Face, error)
	CompareFaces(ctx context.Context, face1, face2 []byte) (float64, error)
}

// OCRResult holds the result of optical character recognition on an image.
type OCRResult struct {
	Text     string     `json:"text"`
	Blocks   []TextBlock `json:"blocks"`
	Language string     `json:"language"`
}

// TextBlock represents a block of detected text with its bounding box.
type TextBlock struct {
	Text       string    `json:"text"`
	Confidence float64   `json:"confidence"`
	BoundingBox []Vertex `json:"bounding_box"`
}

// Vertex represents a point in 2D space.
type Vertex struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Label represents an image classification label.
type Label struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

// Face represents a detected face in an image.
type Face struct {
	Confidence    float64  `json:"confidence"`
	BoundingBox   []Vertex `json:"bounding_box"`
	Joy           string   `json:"joy"`
	Sorrow        string   `json:"sorrow"`
	Anger         string   `json:"anger"`
	Surprise      string   `json:"surprise"`
	Landmarks     []FaceLandmark `json:"landmarks"`
}

// FaceLandmark represents a specific facial feature point.
type FaceLandmark struct {
	Type     string  `json:"type"`
	Position Point3D `json:"position"`
}

// Point3D represents a point in 3D space.
type Point3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// GoogleVisionClient provides image analysis via the Google Cloud Vision API.
type GoogleVisionClient struct {
	credentialsJSON string
	apiKey          string
	httpClient      *http.Client
}

// NewGoogleVisionClient creates a new Google Cloud Vision client.
func NewGoogleVisionClient(credentialsJSON string) *GoogleVisionClient {
	return &GoogleVisionClient{
		credentialsJSON: credentialsJSON,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// visionRequest is the request body for the Google Vision API.
type visionRequest struct {
	Requests []visionAnnotateRequest `json:"requests"`
}

type visionAnnotateRequest struct {
	Image    visionImage    `json:"image"`
	Features []visionFeature `json:"features"`
}

type visionImage struct {
	Content string `json:"content"` // base64-encoded image
}

type visionFeature struct {
	Type       string `json:"type"`
	MaxResults int    `json:"maxResults,omitempty"`
}

// visionResponse is the response body from the Google Vision API.
type visionResponse struct {
	Responses []visionAnnotateResponse `json:"responses"`
}

type visionAnnotateResponse struct {
	TextAnnotations       []visionTextAnnotation `json:"textAnnotations,omitempty"`
	FullTextAnnotation    *visionFullText        `json:"fullTextAnnotation,omitempty"`
	LabelAnnotations      []visionLabel          `json:"labelAnnotations,omitempty"`
	FaceAnnotations       []visionFace           `json:"faceAnnotations,omitempty"`
	Error                 *visionError           `json:"error,omitempty"`
}

type visionTextAnnotation struct {
	Description string `json:"description"`
	Locale      string `json:"locale"`
	BoundingPoly struct {
		Vertices []Vertex `json:"vertices"`
	} `json:"boundingPoly"`
}

type visionFullText struct {
	Text  string `json:"text"`
	Pages []struct {
		Property struct {
			DetectedLanguages []struct {
				LanguageCode string  `json:"languageCode"`
				Confidence   float64 `json:"confidence"`
			} `json:"detectedLanguages"`
		} `json:"property"`
		Blocks []struct {
			Paragraphs []struct {
				Words []struct {
					Symbols []struct {
						Text string `json:"text"`
					} `json:"symbols"`
				} `json:"words"`
			} `json:"paragraphs"`
			Confidence float64 `json:"confidence"`
			BoundingBox struct {
				Vertices []Vertex `json:"vertices"`
			} `json:"boundingBox"`
		} `json:"blocks"`
	} `json:"pages"`
}

type visionLabel struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	MID         string  `json:"mid"`
}

type visionFace struct {
	DetectionConfidence float64 `json:"detectionConfidence"`
	BoundingPoly        struct {
		Vertices []Vertex `json:"vertices"`
	} `json:"boundingPoly"`
	JoyLikelihood      string `json:"joyLikelihood"`
	SorrowLikelihood   string `json:"sorrowLikelihood"`
	AngerLikelihood    string `json:"angerLikelihood"`
	SurpriseLikelihood string `json:"surpriseLikelihood"`
	Landmarks          []struct {
		Type     string `json:"type"`
		Position struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"position"`
	} `json:"landmarks"`
}

type visionError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DetectText performs OCR on the provided image data, useful for KYC document
// verification.
func (g *GoogleVisionClient) DetectText(ctx context.Context, imageData []byte) (*OCRResult, error) {
	encoded := base64.StdEncoding.EncodeToString(imageData)

	reqBody := visionRequest{
		Requests: []visionAnnotateRequest{
			{
				Image: visionImage{Content: encoded},
				Features: []visionFeature{
					{Type: "TEXT_DETECTION"},
					{Type: "DOCUMENT_TEXT_DETECTION"},
				},
			},
		},
	}

	resp, err := g.doRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("google vision detect text: %w", err)
	}

	if len(resp.Responses) == 0 {
		return &OCRResult{}, nil
	}

	annotateResp := resp.Responses[0]
	if annotateResp.Error != nil {
		return nil, fmt.Errorf("google vision API error: %s", annotateResp.Error.Message)
	}

	result := &OCRResult{}

	if annotateResp.FullTextAnnotation != nil {
		result.Text = annotateResp.FullTextAnnotation.Text

		if len(annotateResp.FullTextAnnotation.Pages) > 0 {
			page := annotateResp.FullTextAnnotation.Pages[0]
			if len(page.Property.DetectedLanguages) > 0 {
				result.Language = page.Property.DetectedLanguages[0].LanguageCode
			}

			for _, block := range page.Blocks {
				tb := TextBlock{
					Confidence: block.Confidence,
				}
				if len(block.BoundingBox.Vertices) > 0 {
					tb.BoundingBox = block.BoundingBox.Vertices
				}
				// Reconstruct text from paragraphs/words/symbols
				var blockText string
				for _, para := range block.Paragraphs {
					for _, word := range para.Words {
						for _, sym := range word.Symbols {
							blockText += sym.Text
						}
						blockText += " "
					}
				}
				tb.Text = blockText
				result.Blocks = append(result.Blocks, tb)
			}
		}
	} else if len(annotateResp.TextAnnotations) > 0 {
		result.Text = annotateResp.TextAnnotations[0].Description
		result.Language = annotateResp.TextAnnotations[0].Locale
	}

	log.Debug().
		Str("language", result.Language).
		Int("blocks", len(result.Blocks)).
		Msg("google vision OCR completed")

	return result, nil
}

// DetectLabels classifies the content of an image.
func (g *GoogleVisionClient) DetectLabels(ctx context.Context, imageData []byte) ([]Label, error) {
	encoded := base64.StdEncoding.EncodeToString(imageData)

	reqBody := visionRequest{
		Requests: []visionAnnotateRequest{
			{
				Image: visionImage{Content: encoded},
				Features: []visionFeature{
					{Type: "LABEL_DETECTION", MaxResults: 20},
				},
			},
		},
	}

	resp, err := g.doRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("google vision detect labels: %w", err)
	}

	if len(resp.Responses) == 0 {
		return nil, nil
	}

	annotateResp := resp.Responses[0]
	if annotateResp.Error != nil {
		return nil, fmt.Errorf("google vision API error: %s", annotateResp.Error.Message)
	}

	labels := make([]Label, len(annotateResp.LabelAnnotations))
	for i, la := range annotateResp.LabelAnnotations {
		labels[i] = Label{
			Description: la.Description,
			Score:       la.Score,
		}
	}

	log.Debug().Int("labels", len(labels)).Msg("google vision label detection completed")

	return labels, nil
}

// DetectFaces detects faces in the provided image, useful for selfie matching
// during KYC.
func (g *GoogleVisionClient) DetectFaces(ctx context.Context, imageData []byte) ([]Face, error) {
	encoded := base64.StdEncoding.EncodeToString(imageData)

	reqBody := visionRequest{
		Requests: []visionAnnotateRequest{
			{
				Image: visionImage{Content: encoded},
				Features: []visionFeature{
					{Type: "FACE_DETECTION", MaxResults: 10},
				},
			},
		},
	}

	resp, err := g.doRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("google vision detect faces: %w", err)
	}

	if len(resp.Responses) == 0 {
		return nil, nil
	}

	annotateResp := resp.Responses[0]
	if annotateResp.Error != nil {
		return nil, fmt.Errorf("google vision API error: %s", annotateResp.Error.Message)
	}

	faces := make([]Face, len(annotateResp.FaceAnnotations))
	for i, fa := range annotateResp.FaceAnnotations {
		face := Face{
			Confidence:  fa.DetectionConfidence,
			BoundingBox: fa.BoundingPoly.Vertices,
			Joy:         fa.JoyLikelihood,
			Sorrow:      fa.SorrowLikelihood,
			Anger:       fa.AngerLikelihood,
			Surprise:    fa.SurpriseLikelihood,
		}

		for _, lm := range fa.Landmarks {
			face.Landmarks = append(face.Landmarks, FaceLandmark{
				Type: lm.Type,
				Position: Point3D{
					X: lm.Position.X,
					Y: lm.Position.Y,
					Z: lm.Position.Z,
				},
			})
		}

		faces[i] = face
	}

	log.Debug().Int("faces", len(faces)).Msg("google vision face detection completed")

	return faces, nil
}

// CompareFaces compares two face images and returns a similarity score between 0
// and 1. This uses facial landmark positions to compute a normalized distance.
func (g *GoogleVisionClient) CompareFaces(ctx context.Context, face1, face2 []byte) (float64, error) {
	faces1, err := g.DetectFaces(ctx, face1)
	if err != nil {
		return 0, fmt.Errorf("compare faces - detect face 1: %w", err)
	}
	if len(faces1) == 0 {
		return 0, fmt.Errorf("compare faces: no face detected in first image")
	}

	faces2, err := g.DetectFaces(ctx, face2)
	if err != nil {
		return 0, fmt.Errorf("compare faces - detect face 2: %w", err)
	}
	if len(faces2) == 0 {
		return 0, fmt.Errorf("compare faces: no face detected in second image")
	}

	// Compare using facial landmark positions.
	// This is a simplified approach; in production you would use a dedicated
	// face-comparison service (e.g., AWS Rekognition, Azure Face API).
	score := computeLandmarkSimilarity(faces1[0].Landmarks, faces2[0].Landmarks)

	log.Debug().Float64("similarity", score).Msg("face comparison completed")

	return score, nil
}

// computeLandmarkSimilarity computes a similarity score from facial landmarks.
func computeLandmarkSimilarity(lm1, lm2 []FaceLandmark) float64 {
	if len(lm1) == 0 || len(lm2) == 0 {
		return 0
	}

	// Build lookup maps by landmark type.
	map1 := make(map[string]Point3D)
	for _, l := range lm1 {
		map1[l.Type] = l.Position
	}

	map2 := make(map[string]Point3D)
	for _, l := range lm2 {
		map2[l.Type] = l.Position
	}

	// Compute normalized euclidean distance across matching landmarks.
	var totalDist float64
	var count int
	for typ, p1 := range map1 {
		p2, ok := map2[typ]
		if !ok {
			continue
		}
		dx := p1.X - p2.X
		dy := p1.Y - p2.Y
		dz := p1.Z - p2.Z
		totalDist += math.Sqrt(dx*dx + dy*dy + dz*dz)
		count++
	}

	if count == 0 {
		return 0
	}

	avgDist := totalDist / float64(count)

	// Convert distance to a 0-1 similarity score using an exponential decay.
	// Lower distance = higher similarity.
	similarity := math.Exp(-avgDist / 100.0)

	return math.Min(1.0, math.Max(0.0, similarity))
}

// doRequest performs the HTTP request to the Google Vision API.
func (g *GoogleVisionClient) doRequest(ctx context.Context, reqBody visionRequest) (*visionResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal vision request: %w", err)
	}

	url := googleVisionEndpoint
	if g.apiKey != "" {
		url += "?key=" + g.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create vision request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read vision response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google vision API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result visionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal vision response: %w", err)
	}

	return &result, nil
}
