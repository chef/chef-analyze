package reporting

const (
	// This is a ruby formatter that works with Cookstyle to output
	// the results as a simple set of values: num-errors-found num-correctable-errors-found
	correctableCountCookstyleFormatterRb = `
require 'json'

module Local
  # A simple formatter that outputs the total number of violations
  # found, and the total number of those that are auto-correctable.
  class CorrectableCountFormatter < RuboCop::Formatter::BaseFormatter
    def initialize(output, options = {})
      super
      @offense_count = 0
      @correctable_count = 0
      @registry = RuboCop::Cop::Registry.new(RuboCop::Cop::Cop.all)
    end

    def file_finished(_file, offenses)
      @offense_count += offenses.count
      offenses.each do |offense|
        # If a cop has an autocorrect method (inherited or otherwise) then it can
        # be autocorrected:
        @correctable_count += 1 if @registry.find_by_cop_name(offense.cop_name).method_defined?(:autocorrect)
      end
    end

    def finished(_inspected_files)
      output.write "#{@offense_count} #{@correctable_count}"
    end
  end
end
`
)
