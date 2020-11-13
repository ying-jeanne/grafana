import React from 'react';
import uPlot from 'uplot';
import { DataFrame, FieldColor, TimeRange, TimeZone } from '@grafana/data';

export type NullValuesMode = 'null' | 'connected' | 'asZero';

export enum AxisSide {
  Top,
  Right,
  Bottom,
  Left,
}

interface AxisConfig {
  label: string;
  side: AxisSide;
  grid: boolean;
  width: number;
}

interface LineConfig {
  show: boolean;
  width: number;
  color: FieldColor;
}
interface PointConfig {
  show: boolean;
  radius: number;
}
interface BarsConfig {
  show: boolean;
}
interface FillConfig {
  alpha: number;
}

export interface GraphCustomFieldConfig {
  axis: AxisConfig;
  line: LineConfig;
  points: PointConfig;
  bars: BarsConfig;
  fill: FillConfig;
  nullValues: NullValuesMode;
}

export type PlotPlugin = {
  id: string;
  /** can mutate provided opts as necessary */
  opts?: (self: uPlot, opts: uPlot.Options) => void;
  hooks: uPlot.PluginHooks;
};

export interface PlotPluginProps {
  id: string;
}

export interface PlotProps {
  data: DataFrame;
  timeRange: TimeRange;
  timeZone: TimeZone;
  width: number;
  height: number;
  children?: React.ReactNode | React.ReactNode[];
  /** Callback performed when uPlot data is updated */
  onDataUpdate?: (data: uPlot.AlignedData) => {};
  /** Callback performed when uPlot is (re)initialized */
  onPlotInit?: () => {};
}

export class UPlotChartConfig {
  private axes: UPlotChartAxis[] = [];
  private scales: UPlotChartScale[] = [];
  private series: UPlotChartSeries[] = [];

  public addAxis(props: UPlotChartAxisProps) {
    this.axes.push(new UPlotChartAxis(props));
  }

  public addSeries(props: UPlotSeriesProps) {
    this.series.push(new UPlotChartSeries(props));
  }

  public addScale(props: UPlotScaleProps) {
    this.scales.push(new UPlotChartScale(props));
  }

  public hasScale(key: string) {
    return false;
  }

  getConfig() {}
}

export interface UPlotChartAxisProps {
  scaleKey: string;
  label?: string;
  show?: boolean;
  size?: number;
  stroke?: string;
  grid?: boolean;
  formatValue?: (v: any) => string;
  values?: any;
  isTime?: boolean;
  timeZone?: TimeZone;
}

export class UPlotChartAxis {
  constructor(private props: UPlotChartAxisProps) {}

  getConfig(): uPlot.Axis {
    return {};
  }
}

export interface UPlotChartAxisProps {
  scaleKey: string;
  label?: string;
  show?: boolean;
  size?: number;
  stroke?: string;
  grid?: boolean;
  formatValue?: (v: any) => string;
  values?: any;
  isTime?: boolean;
  timeZone?: TimeZone;
  side: AxisSide;
}

export interface UPlotScaleProps {
  scaleKey: string;
  isTime?: boolean;
}

export class UPlotChartScale {
  constructor(private props: UPlotScaleProps) {}

  getConfig(): uPlot.Scale {
    return {};
  }
}

interface UPlotSeriesProps {
  lines?: boolean;
  lineWidth?: number;
  areaFill?: number;
  points?: boolean;
  pointSize?: number;
  color?: string;
  scaleKey: string;
}

export class UPlotChartSeries {
  constructor(private props: UPlotSeriesProps) {}

  getConfig(): uPlot.Scale {
    return {};
  }
}
