class Task < ApplicationRecord

    enum :status, { pending: 0, in_progress: 1, completed: 2 }

    scope :completed, -> { where(status: :completed) }
    scope :in_progress, -> { where(status: :in_progress) }
    scope :pending, -> { where(status: :pending) }
    scope :not_completed, -> { where.not(status: :completed) }

    def self.human_attribute_enum_value(attr_name, value)
        return if value.blank?
        human_attribute_name("#{attr_name}.#{value}")
    end

    def human_attribute_enum(attr_name)
        self.class.human_attribute_enum_value(attr_name, self.send("#{attr_name}"))
    end
end
