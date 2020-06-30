app.controller("dashboardController",["$scope","$rootScope","dashboardService","requestService","modalService","errorService",
function($scope,$rootScope,dashboardService,requestService,modalService,errorService){
    var self = this;
    var url = window.location.hash.split("/")[1];
    $scope.account = {};
    $scope.leftbar = {};
    $scope.job_list = [];
    $scope.new_password = {
        old_password: "",
        new_password: "",
        new_password_verification: ""
    }
    $scope.list = [
        {
            "name":"all_jobs",
            "fontawesome": "fas fa-list-ol"
        },
        {
            "name":"create_job",
            "fontawesome": "fas fa-plus"
        }
    ];
    console.dir($scope);
    $rootScope.$on("account_action", function(e, list) {
        list.forEach(function(value){
            $scope.account[value.name] = url == value.name;
        })
    });
    $scope.list.forEach(function(value){
        $scope.leftbar[value.name] = url == value.name;
    });
    $scope.change_password = function (){
        console.log("change_password");
        dashboardService.change_password($scope.new_password).then(function(response){
            if(response.data.code == 200){
                data = {
                    title: "success",
                    buttons: modalService.buttons["ok"]
                }
                modalService.open_dialog(modalService.modals.popup,data,broadcast_logout);
            }else{
                data = errorService.handle_error(response.data.code);
                modalService.open_popup(modalService.modals.popup,data);
            }
        }).catch(function(response){
            var error = errorService.handle_error(response.data.code);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });
    }
    function broadcast_logout(){
        $rootScope.$emit("logout");
    }
}]);