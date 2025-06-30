class AddColumnToTask < ActiveRecord::Migration[8.0]
  def change
    add_column :tasks, :name, :string, null: false
    add_column :tasks, :context, :text, null: false
    add_column :tasks, :status, :integer, null: false
    add_column :tasks, :deadline, :datetime, null: true
  end
end
