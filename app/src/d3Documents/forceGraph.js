import * as d3 from 'd3';
import { isFunction } from './util';

function drag(simulation) {
  function dragstarted(event) {
    if (!event.active) simulation.alphaTarget(0.3).restart();
    event.subject.fx = event.subject.x;
    event.subject.fy = event.subject.y;
  }

  function dragged(event) {
    event.subject.fx = event.x;
    event.subject.fy = event.y;
  }

  function dragended(event) {
    if (!event.active) simulation.alphaTarget(0);
    event.subject.fx = null;
    event.subject.fy = null;
  }

  return d3.drag()
    .on('start', dragstarted)
    .on('drag', dragged)
    .on('end', dragended);
}

export default class ForceGraph {
  constructor(width, height, svgId) {
    this.svgId = svgId;
    this.width = width;
    this.height = height;

    // add attributes to root svg
    this.rootSvg = d3.select(`#${this.svgId}`).attr('viewBox', `0 0 ${this.width} ${this.height}`);
    this.rootGroup = this.rootSvg.append('g').attr('class', 'root-group');

    this.lineGroup = this.rootGroup.append('g')
      .attr('stroke', '#999')
      .attr('stroke-opacity', 0.6);

    this.nodeGroup = this.rootGroup
      .append('g')
      .attr('stroke', '#fff')
      .attr('stroke-width', '#C2C2C2');

    // add zoom and drag
    this.zoom = d3.zoom()
      .on('zoom', (event) => {
        this.rootGroup.attr('transform', event.transform);
      })
      .scaleExtent([0.25, 8]);
    this.rootSvg.call(this.zoom);

    // click
    this.clickCallBack = null;

    // add arrow defintion
    this.rootSvg
      .append('svg:defs')
      .append('svg:marker')
      .attr('id', 'arrowhead')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 22)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('svg:path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', '#999');
  }

  draw(nodes, links, colorMap) {
    if (this.clickCallBack === null) {
      throw new Error('click call back is not set');
    }

    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(links).id((d) => d.txhash))
      .force('charge', d3.forceManyBody().strength(-80))
      .force('x', d3.forceX(this.width / 2).strength(0.04))
      .force('y', d3.forceY(this.height / 2).strength(0.04))
      .force('radial', d3.forceRadial(240, this.width / 2, this.height / 2))
      .force('center', d3.forceCenter(this.width / 2, this.height / 2));

    const link = this.lineGroup
      .selectAll('.arrow')
      .data(links, (d) => `${d.source}${d.target}`)
      .join('line')
      .attr('class', 'arrow')
      .attr('marker-end', 'url(#arrowhead)')
      .attr('stroke-width', 1);

    const node = this.nodeGroup
      .selectAll('circle')
      .data(nodes, (d) => d.txhash)
      .join('circle')
      .attr('r', 7)
      .attr('fill', (d) => colorMap.get(d.privacytype))
      .on('click', (e, d) => { this.clickCallBack(d); })
      .on('mouseover', function mouseOver() {
        d3.select(this).attr('r', 10).classed('nodeMouseOver', true);
      })
      .on('mouseout', function mouseOut() {
        d3.select(this).attr('r', 7).classed('nodeMouseOver', false);
      })
      .call(drag(simulation));

    node.append('title')
      .text((d) => `${d.txhash}\n${d.privacytype}`);

    simulation.on('tick', () => {
      link
        .attr('x1', (d) => d.source.x)
        .attr('y1', (d) => d.source.y)
        .attr('x2', (d) => d.target.x)
        .attr('y2', (d) => d.target.y);

      node
        .attr('cx', (d) => d.x)
        .attr('cy', (d) => d.y);
    });
  }

  setClickHandler(callback) {
    if (!isFunction(callback)) {
      return false;
    }

    this.clickCallBack = callback;
    return true;
  }
}
