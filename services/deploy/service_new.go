package deploy

type service struct {
}

func NEW() *service {
	return &service{}
}

func (s *service) NEW_DEPLOYMENT(deployment *Deployment_New) 
{

	// add this deployment to DB

	for _, service := deployment.Services {
		// docker compose -> make services automatically
		// if no docker compoes -> make services manually
	}
	
}

func (s *service) NEW_SERVICE() 
{
}
