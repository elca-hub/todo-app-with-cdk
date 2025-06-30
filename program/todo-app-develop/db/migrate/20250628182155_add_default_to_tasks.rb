class AddDefaultToTasks < ActiveRecord::Migration[8.0]
  def change
    change_column_default :tasks, :status, from: nil, to: 0
  end
end
