package services

type Predictor interface {
	Predict(PredictionRequest) (*PredictionResponse, error)
}

// compile-time assertions
var _ Predictor = (*PredictionService)(nil)
var _ Predictor = (*PredictionServiceV2)(nil)
