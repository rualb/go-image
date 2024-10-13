// Package service app services
package service

func mustCreateRepository(appService AppService) {

	_ = appService.Repository()

	mustInitRepositoryMasterData(appService)
}

func mustInitRepositoryMasterData(appService AppService) {
	_ = appService.Repository()

}
