# Multistage builder calibration
The historical Kaniko 1.5.1 task default compiled the application but failed during stage cleanup at /product_uuid (device busy). It did not publish an image and no recovery arm executed. Retained as environment calibration; evaluated runs use a pinned contemporary builder via the public Task BUILDER_IMAGE parameter.
