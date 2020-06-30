app
.controller("createjobController",["$scope","dashboardService","requestService","modalService","errorService",
function($scope,dashboardService,requestService,modalService,errorService){
    $scope.selected_task = null;
    $scope.tasks = [];
    $scope.public_date = "";
    $scope.exp_date = "";
    $scope.job = {
        tasks:[]
    };

    $scope.add_task = function() {
        available_services();
        var data = {
            title: "add_task",
            buttons: modalService.buttons["ok/cancel"]
        }
        modalService.open_dialog(modalService.modals["task_list"],data,add_task,$scope,"modalController")
    }

    $scope.post_job = function(){
        var publication_date = get_epoch($scope.public_date);
        var expiration_date = get_epoch($scope.exp_date);

        console.log("Create job with pub=" + publication_date + " and exp date=" + expiration_date);

        dashboardService.create_job(publication_date, expiration_date, $scope.job.tasks).
            then(function(response){
                if(response.data.code == 200)
                {
                    data= {
                        title : "success",
                        buttons: modalService.buttons["ok"]
                    }
                    clear_inputs();
                }else{
                    data = errorService.handle_error(response.data.code);
                }
                modalService.open_popup(modalService.modals.popup,data);
            }).catch(function(response){
                var error = errorService.handle_error(response.data.code);
                error.message = response.data.message;
                modalService.open_popup(modalService.modals.popup,error);
            });
    }

    $scope.remove_task = function(index){
        var data = {
            title: "delete_task_confirmation",
            buttons: modalService.buttons["yes/no"]
        }
        modalService.open_dialog(modalService.modals.dialog,data,function(ok){
            if(ok == 1){
                $scope.tasks.splice(index,1);
                $scope.job.tasks.push(index,1);
            }
        });
    }

    $scope.clear_linked_input = function(task_index,key,linked){
        if(linked == false){
            delete $scope.job.tasks[task_index].linked_input[key];
        }else if(linked == true){
            if($scope.job.tasks[task_index].input.hasOwnProperty(key)){
                delete $scope.job.tasks[task_index].input[key];
           }
        }
    }
    function available_services () {
        dashboardService.get_services()
        .then(function(response){
            console.dir(response);
            if(response.data.code == 200){
                $scope.services = response.data.services;
            }else{
                throw response;
            }
        }).catch(function(response){
            var error = errorService.handle_error(response.data.code);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });;
    }
    function add_task(ok){
        if(ok == 1){
            $scope.tasks.push($scope.selected_task);
            var task = {
                task_id: $scope.selected_task.id,
                input : {},
                linked_input: {}
            }
            $scope.selected_task = null;
            $scope.job.tasks.push(task);
        }else{
            $scope.selected_task = null;
        }
    }

    function get_epoch(date){
        var parts = date.split("-");
        var dt = new Date(parseInt(parts[2], 10),
                        parseInt(parts[0], 10) - 1,
                        parseInt(parts[1], 10));
        return dt/1000;
        // The bellow doesn't work on firefox
        //return new Date(date).valueOf()/1000;
    }

    function clear_inputs(){
        $scope.public_date = "";
        $scope.exp_date = "";
        $scope.selected_task = null;
        $scope.tasks = [];
        $scope.job = {
            tasks:[]
        };
    }
}])