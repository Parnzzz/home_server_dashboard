import 'package:flutter/material.dart';

class TimeRangeSelector extends StatelessWidget {
  final String selectedRange;
  final ValueChanged<String> onChanged;

  const TimeRangeSelector({
    super.key,
    required this.selectedRange,
    required this.onChanged,
  });

  static const List<String> ranges = [
    '1m',
    '5m',
    '15m',
    '1h',
  ];

  @override
  Widget build(BuildContext context) {
    return SegmentedButton<String>(
      segments: const [
        ButtonSegment(
          value: '1m',
          label: Text('1 Min'),
        ),
        ButtonSegment(
          value: '5m',
          label: Text('5 Min'),
        ),
        ButtonSegment(
          value: '15m',
          label: Text('15 Min'),
        ),
        ButtonSegment(
          value: '1h',
          label: Text('1 Hour'),
        ),
      ],
      selected: {selectedRange},
      onSelectionChanged: (selection) {
        onChanged(selection.first);
      },
      showSelectedIcon: false,
    );
  }
}