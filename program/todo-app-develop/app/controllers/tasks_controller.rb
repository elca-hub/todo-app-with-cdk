class TasksController < ApplicationController

    before_action :set_task, only: [:show, :edit, :update, :destroy]

    def index()
        @tasks = Task.all
        @completed_tasks = Task.completed
        @not_completed_tasks = Task.not_completed
    end

    def show()
    end

    def new()
        @task = Task.new
    end

    def create()
        @task = Task.new(task_params)
        if @task.save
            redirect_to tasks_path, notice: "作成完了"
        else
            render :new, alert: "作成失敗"
        end
    end

    def edit()
    end

    def update()
        if @task.update(task_params)
            redirect_to @task, notice: "更新完了"
        else
            render :edit, alert: "更新失敗"
        end
    end

    def destroy()
        @task.destroy
        redirect_to tasks_path, notice: "削除完了"
    end

    private

    def set_task
        @task = Task.find(params[:id])
    end

    def task_params
        params.require(:task).permit(:name, :context, :status, :deadline)
    end
end
