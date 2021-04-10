package markers

import "backend/model-parser/model"

func textParameter(name string, caption string, required bool) model.MarkerParameter {
	return model.MarkerParameter{
		Name:          name,
		Caption:       caption,
		ParameterType: model.TEXT,
		Required:      required,
	}
}

func inspectorParameter(name string, caption string, required bool) model.MarkerParameter {
	return model.MarkerParameter{
		Name:          name,
		Caption:       caption,
		ParameterType: model.INSPECTOR,
		Required:      required,
	}
}

func makeParameters(parameters ...model.MarkerParameter) []model.MarkerParameter {
	return parameters
}
