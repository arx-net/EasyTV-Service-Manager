app 
.controller("modalController",["$scope","modalService","errorService",function($scope,modalService,errorService){
    $scope.close_dialog = function(ok) {
        if(ok == 1){
            if($scope.$parent.selected_task != null){
                $scope.closeThisDialog(ok);
            }else{
                var error = errorService.handle_error("client-102");
                modalService.open_popup(modalService.modals.popup,error);
                $scope.closeThisDialog(0);
            }
        }else{
            $scope.closeThisDialog(0)
        }
    }
}])