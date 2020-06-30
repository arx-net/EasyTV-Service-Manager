app
.controller("adminDashboardController",["$scope","$rootScope","adminDashboardService","modalService","errorService",
function($scope,$rootScope,adminDashboardService,modalService,errorService){
    var self = this;
    var url = window.location.hash.split("/")[1];
    $scope.account = {};
    $scope.leftbar = {};
    $scope.new_password = {
        old_password: "",
        new_password: "",
        new_password_verification: ""
    }
    $scope.name = "";
    $scope.username = "";
    $scope.email = "";
    $scope.description = "";
    $scope.response = ""; 
    $scope.list = [
        {
            "name":"register_user",
            "fontawesome": "fas fa-plus-square"
        },
        {
            "name":"create_service",
            "fontawesome": "fas fa-plus"
        },
        {
            "name":"get_services",
            "fontawesome": "fas fa-list-ol"
        }
    ]
    $rootScope.$on("account_action", function(e, list) {
        list.forEach(function(value){
            $scope.account[value.name] = url == value.name;
        })
    });
    $scope.list.forEach(function(value){
        $scope.leftbar[value.name] = url == value.name;
    })
    $scope.submit_register = function(){
        adminDashboardService.register_user($scope.name,$scope.username,$scope.email)
        .then(function(response){
            if(response.data.code == 200){
                var data = {
                    title: "success",
                    message: "password",
                    param: ": " + response.data.content_owner_password,
                    buttons: modalService.buttons["ok"]
                }
                clear_inputs();
            }else{
                data = errorService.handle_error(response.data.code);
            } 
            modalService.open_popup(modalService.modals.popup,data)
        }).catch(function(response){
            var error = errorService.handle_error(response.data.code);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });
    }
    $scope.submit_create_service = function(){
        adminDashboardService.create_service($scope.name,$scope.description)
        .then(function(response){
            data = {
                buttons: modalService.buttons["ok"]
            }
            if(response.data.code == 200){
                data = {
                    title: "success",
                    message: "It was registered with the api key: " + response.data.api_key,
                    buttons: modalService.buttons["ok"]
                }
                clear_inputs();
            }else{
                data = errorService.handle_error(response.data.code);
            }
            modalService.open_popup(modalService.modals.popup, data, 700)
        }).catch(function(response){
            console.dir(response);
            var error = errorService.handle_error(response.data.code);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });
    }
    $scope.request_service_availability_change = function(service){
        if(service.enabled == false){
            data = {
                param: service.name,
                title : 'service_disable_confirmation_title',
                message: 'service_disable_confirmation_msg',
                buttons: modalService.buttons["ok/cancel"]
            }
        }else{
            data = {
                param: service.name,
                title : 'service_enable_confirmation_title',
                message: 'service_enable_confirmation_msg',
                buttons: modalService.buttons["ok/cancel"]
            }
        }
        modalService.open_dialog(modalService.modals.dialog,data,(ok)=>{
            if(ok == 1){
                $scope.submit_service_availability_change(service);
            }else if(ok == 0){
                service.enabled = !service.enabled;
            }else{
                $scope.$apply(function(){
                    service.enabled = !service.enabled;
                })
            }
        });
    }
    $scope.submit_service_availability_change = function(service){
        adminDashboardService.service_availability_change(service)
        .then(function(response){
            if(response.data.code == 200){
                data = {
                    title: "success",
                    buttons: modalService.buttons["ok"]
                }
            }else{
                data = errorService.handle_error(response.data.code);
            }
            modalService.open_popup(modalService.modals.popup,data);
        }).catch(function(response){
            var error = errorService.handle_error(response.data.code);
            console.dir(response);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });
    }
    $scope.get_services = function(){
        adminDashboardService.get_services()
        .then(function(response){
            if(response.data.code == 200){
                $scope.response = response.data.description;
                $scope.services = response.data.services;
                $(document).ready(function(){
                    $('#services').DataTable();
                })
            }else{
                throw response;
            }
        }).catch(function(response){
            var error = errorService.handle_error(response.data.code);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });
    }
    $scope.change_password = function (){
        adminDashboardService.change_password($scope.new_password).then(function(response){
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
    if(url == "get_services"){
        $scope.get_services();
    }
    function broadcast_logout(){
        $rootScope.$emit("logout");
    }
    function clear_inputs(){
        $scope.name = "";
        $scope.username = "";
        $scope.email = "";
        $scope.description = "";
    }
}]);