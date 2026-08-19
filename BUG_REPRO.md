# Bug reproduction

Wafer selection, sorting, annotation, and reclassification reuse the source backing array and nested mutable values. Derived results therefore rewrite the original wafer collection.

Run the targeted wafer test. It reports that the source slot order, tags, measurements, or timestamp changed after a derived operation.
