package api

import "cloudsweep/model"

func getResponse500() *model.Response500 {
	return &model.Response500{
		StatusCode: 500,
	}
}

func getResponse400() *model.Response400 {
	return &model.Response400{
		StatusCode: 400,
	}
}

func getResponse409() *model.Response409 {
	return &model.Response409{
		StatusCode: 409,
	}
}

func getResponse200() *model.Response200 {
	return &model.Response200{
		StatusCode: 200,
	}
}

func getResponse207() *model.Response207 {
	return &model.Response207{
		StatusCode: 207,
	}
}

func getResponse404() *model.Response404 {
	return &model.Response404{
		Error:      "Requested Object Not Found or You are not Authorized to access",
		StatusCode: 404,
	}
}

func getResponse201() *model.Response201 {
	return &model.Response201{
		StatusCode: 201,
	}
}
