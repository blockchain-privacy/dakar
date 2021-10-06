import * as d3 from 'd3';
import { Tree } from './tree';

export default class ClusterTree extends Tree {
  constructor(width) {
    super(width);

    this.drawResetClick = () => {
      this.rootSvg.selectAll('.rect').classed('clicked', false);
    };

    this.drawClicked = (node) => {
      node.select('.rect').classed('clicked', true);
    };

    this.stratify = d3.stratify()
      .id((d) => d.uid)
      .parentId((d) => {
        if (d.cluster_parent === null || d.cluster_parent === undefined) return null;
        return d.cluster_parent;
      });

    this.drawTree = (data) => {
      this.drawLinks(this.rootGroup, data);
      this.drawNodes(this.rootGroup, data);
    };
  }

  drawLinks(group, nodeData) {
    // adds the links between the nodes
    const t = d3.transition()
      .duration(300)
      .ease(d3.easeLinear);
    group.selectAll('.link')
      .data(nodeData.descendants().slice(1), (d) => d.data.data.uid)
      .join('path')
      .attr('class', 'link')
      .transition(t)
      .attr('stroke-opacity', 1)
      .attr('d', (d) => `M${d.y - this.rectWidth / 2},${d.x
      }C${(d.y + d.parent.y) / 2},${d.x
      } ${(d.y + d.parent.y) / 2},${d.parent.x
      } ${d.parent.y + this.rectWidth / 2},${d.parent.x}`);
  }

  drawRect(rootElement) {
    const textAreaHeight = 50;
    const textPadding = 0;
    const rectHeight = textAreaHeight + 2 * textPadding;
    const borderRadius = 5;
    const strokeWidth = 2;
    const textHeight = 10;

    rootElement
      .append('rect')
      .attr('x', -this.rectWidth / 2)
      .attr('y', -rectHeight / 2)
      .attr('class', 'rect')
      .attr('width', this.rectWidth)
      .attr('height', rectHeight)
      .attr('rx', borderRadius)
      .attr('ry', borderRadius)
      .attr('stroke-width', strokeWidth)
      .attr('stroke-opacity', 1);

    rootElement
      .append('text')
      .attr('fill', 'currentColor')
      .attr('x', () => -this.rectWidth / 2 + strokeWidth + 2)
    // if parameter is not set, position text at center
      .attr('y', (d) => (
        d.data.data.parameter !== undefined
          ? -textAreaHeight / 2 + textHeight * 2 : textHeight / 2))
      .text((d) => `Size:${d.data.data.cluster_address_count}`);
  }

  drawNodes(group, nodeData) {
    const self = this;

    // adds each node as a group
    const t = d3.transition().duration(300).ease(d3.easeLinear);

    return group.selectAll('.node')
      .data(nodeData.descendants(), (d) => d.data.data.uid)
      .join((enter) => {
        const g = enter.append('g')
          .on('click', function click(e) { Tree.nodeClicked(e, self, this); });
          // draw outline and text
        self.drawRect(g);
        return g;
      })
      .attr('opacity', 1)
      .attr('class', (d) => `node${
        d.children ? ' node--internal' : ' node--leaf'}`)
      .transition(t)
      .attr('transform', (d) => `translate(${d.y},${d.x})`);
  }

  drawForce(data) {
    const treeData = this.stratify(data);

    //  assigns the data to a hierarchy using parent-child relationships
    const root = d3.hierarchy(treeData, (d) => d.children);

    const links = root.links();
    const nodes = root.descendants();

    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(links).id((d) => d.id).distance(0).strength(2))
      .force('charge', d3.forceManyBody().strength(-30))
    // .force('r', d3.forceRadial(150, 300, 300).strength(0.7))
      .force('x', d3.forceX().strength(0.005))
      .force('y', d3.forceY().strength(0.005));

    const link = this.rootGroup.append('g')
      .attr('stroke', '#999')
      .attr('stroke-opacity', 0.6)
      .selectAll('line')
      .data(links)
      .join('line');

    const node = this.rootGroup.append('g')
      .attr('fill', '#fff')
      .attr('stroke', '#000')
      .attr('stroke-width', 1.5)
      .selectAll('circle')
      .data(nodes)
      .join('circle')
      .attr('fill', (d) => (d.children && d.children.length > 1 ? 'red' : '#005e7e'))
      .attr('stroke', (d) => (d.children && d.children.length > 1 ? 'red' : '#005e7e'))
      .attr('r', 7);

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
}
