import React, { useMemo } from 'react';
import {
  DataFrame,
  FieldConfig,
  FieldType,
  formattedValueToString,
  getFieldColorModeForField,
  getFieldDisplayName,
  getTimeField,
  TIME_SERIES_TIME_FIELD_NAME,
} from '@grafana/data';
import { alignAndSortDataFramesByFieldName } from './utils';
import { UPlotChart } from '../uPlot/Plot';
import { AxisSide, GraphCustomFieldConfig, PlotProps, UPlotChartConfig } from '../uPlot/types';
import { useTheme } from '../../themes';
import { VizLayout } from '../VizLayout/VizLayout';
import { LegendItem, LegendOptions } from '../Legend/Legend';
import { GraphLegend } from '../Graph/GraphLegend';

const defaultFormatter = (v: any) => (v == null ? '-' : v.toFixed(1));

interface GraphNGProps extends Omit<PlotProps, 'data'> {
  data: DataFrame[];
  legend?: LegendOptions;
}

export const GraphNG: React.FC<GraphNGProps> = ({
  data,
  children,
  width,
  height,
  legend,
  timeRange,
  timeZone,
  ...plotProps
}) => {
  const theme = useTheme();
  const alignedData = useMemo(() => alignAndSortDataFramesByFieldName(data, TIME_SERIES_TIME_FIELD_NAME), [data]);

  if (!alignedData) {
    return (
      <div className="panel-empty">
        <p>No data found in response</p>
      </div>
    );
  }

  const plotConfig = useMemo(() => {
    const plotConfig = new UPlotChartConfig();

    let { timeIndex } = getTimeField(alignedData);
    if (timeIndex === undefined) {
      timeIndex = 0; // assuming first field represents x-domain
      plotConfig.addScale({ scaleKey: 'x' });
    } else {
      plotConfig.addScale({ scaleKey: 'x', isTime: true });
    }

    plotConfig.addAxis({ scaleKey: 'x', isTime: true, side: AxisSide.Bottom, timeZone: timeZone });

    let seriesIdx = 0;

    for (let i = 0; i < alignedData.fields.length; i++) {
      const field = alignedData.fields[i];
      const config = field.config as FieldConfig<GraphCustomFieldConfig>;
      const customConfig = config.custom;

      if (i === timeIndex || field.type !== FieldType.number) {
        continue;
      }

      const fmt = field.display ?? defaultFormatter;
      const scale = config.unit || '__fixed';

      if (!plotConfig.hasScale(scale)) {
        plotConfig.addScale({ scaleKey: scale });
        plotConfig.addAxis({
          scaleKey: scale,
          label: config.custom?.axis?.label,
          size: config.custom?.axis?.width,
          side: config.custom?.axis?.side || AxisSide.Left,
          grid: config.custom?.axis?.grid,
          formatValue: v => formattedValueToString(fmt(v)),
        });
      }

      // need to update field state here because we use a transform to merge framesP
      field.state = { ...field.state, seriesIndex: seriesIdx };

      const colorMode = getFieldColorModeForField(field);
      const seriesColor = colorMode.getCalculator(field, theme)(0, 0);

      plotConfig.addSeries({
        lines: customConfig?.line?.show,
        scaleKey: scale,
        color: seriesColor,
        lineWidth: customConfig?.line.show ? customConfig?.line.width || 1 : 0,
        points: customConfig?.points?.show,
        pointSize: customConfig?.points?.radius,
        areaFill: customConfig?.fill.alpha,
      });

      seriesIdx++;
    }
  }, [alignedData]);

  return (
    <VizLayout width={width} height={height}>
      {(vizWidth: number, vizHeight: number) => (
        <UPlotChart
          data={alignedData}
          width={vizWidth}
          height={vizHeight}
          timeRange={timeRange}
          timeZone={timeZone}
          config={plotConfig}
          {...plotProps}
        >
          {scales}
          {axes}
          {geometries}
          {children}
        </UPlotChart>
      )}
    </VizLayout>
  );
};
